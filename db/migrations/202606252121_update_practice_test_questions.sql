-- migrate:up
DO $$
DECLARE
    r RECORD;
    q jsonb;
    new_q jsonb;
    questions_arr jsonb;
    mcq_obj jsonb;
    answered_index int;
    i int;
BEGIN
    FOR r IN SELECT id, questions FROM practice_tests WHERE questions IS NOT NULL AND jsonb_typeof(questions) = 'array' LOOP
        questions_arr := '[]'::jsonb;

        FOR q IN SELECT * FROM jsonb_array_elements(r.questions) LOOP
            new_q := q - 'questionType';

            -- Rename trueFalseQuestion -> tfq
            IF q ? 'trueFalseQuestion' THEN
                new_q := new_q - 'trueFalseQuestion';
                new_q := jsonb_set(new_q, '{tfq}', q->'trueFalseQuestion');
            END IF;

            -- Transform MCQ: answeredTerm + correctChoiceIndex -> answeredIndex
            IF q ? 'mcq' AND jsonb_typeof(q->'mcq') = 'object' THEN
                mcq_obj := q->'mcq';

                IF (mcq_obj->>'correct') = 'true' AND mcq_obj ? 'correctChoiceIndex' THEN
                    answered_index := (mcq_obj->>'correctChoiceIndex')::int;
                ELSIF (mcq_obj->>'correct') = 'false'
                      AND mcq_obj ? 'answeredTerm' AND jsonb_typeof(mcq_obj->'answeredTerm') = 'object'
                      AND mcq_obj ? 'distractors' AND jsonb_typeof(mcq_obj->'distractors') = 'array' THEN
                    answered_index := -1;
                    FOR i IN 0..jsonb_array_length(mcq_obj->'distractors')-1 LOOP
                        IF mcq_obj->'distractors'->i->>'id' = mcq_obj->'answeredTerm'->>'id' THEN
                            IF i >= (mcq_obj->>'correctChoiceIndex')::int THEN
                                answered_index := i + 1;
                            ELSE
                                answered_index := i;
                            END IF;
                            EXIT;
                        END IF;
                    END LOOP;
                    IF answered_index = -1 THEN
                        answered_index := 0;
                    END IF;
                ELSE
                    answered_index := 0;
                END IF;

                mcq_obj := mcq_obj - 'answeredTerm';
                mcq_obj := jsonb_set(mcq_obj, '{answeredIndex}', to_jsonb(answered_index));

                IF mcq_obj ? 'term' AND jsonb_typeof(mcq_obj->'term') = 'object' THEN
                    mcq_obj := jsonb_set(mcq_obj, '{term}', (mcq_obj->'term') - 'sortOrder');
                END IF;
                IF mcq_obj ? 'distractors' AND jsonb_typeof(mcq_obj->'distractors') = 'array' THEN
                    FOR i IN 0..jsonb_array_length(mcq_obj->'distractors')-1 LOOP
                        mcq_obj := jsonb_set(mcq_obj, array['distractors', i::text], (mcq_obj->'distractors'->i) - 'sortOrder');
                    END LOOP;
                END IF;

                new_q := jsonb_set(new_q, '{mcq}', mcq_obj);
            END IF;

            -- Strip sortOrder from tfq term and distractor
            IF new_q ? 'tfq' AND jsonb_typeof(new_q->'tfq') = 'object' THEN
                IF new_q->'tfq' ? 'term' AND jsonb_typeof(new_q->'tfq'->'term') = 'object' THEN
                    new_q := jsonb_set(new_q, '{tfq,term}', (new_q->'tfq'->'term') - 'sortOrder');
                END IF;
                IF new_q->'tfq' ? 'distractor' AND jsonb_typeof(new_q->'tfq'->'distractor') = 'object' THEN
                    new_q := jsonb_set(new_q, '{tfq,distractor}', (new_q->'tfq'->'distractor') - 'sortOrder');
                END IF;
            END IF;

            -- Strip sortOrder from frq term and ensure answeredString is non-null
            IF new_q ? 'frq' AND jsonb_typeof(new_q->'frq') = 'object' THEN
                IF new_q->'frq' ? 'term' AND jsonb_typeof(new_q->'frq'->'term') = 'object' THEN
                    new_q := jsonb_set(new_q, '{frq,term}', (new_q->'frq'->'term') - 'sortOrder');
                END IF;
                IF (new_q->'frq'->>'answeredString') IS NULL THEN
                    new_q := jsonb_set(new_q, '{frq,answeredString}', to_jsonb(''));
                END IF;
            END IF;

            questions_arr := questions_arr || new_q;
        END LOOP;

        UPDATE practice_tests SET questions = questions_arr WHERE id = r.id;
    END LOOP;
END;
$$;

-- migrate:down
DO $$
DECLARE
    r RECORD;
    q jsonb;
    new_q jsonb;
    questions_arr jsonb;
    question_type text;
    mcq_obj jsonb;
    answered_idx int;
    cci int;
BEGIN
    FOR r IN SELECT id, questions FROM practice_tests WHERE questions IS NOT NULL AND jsonb_typeof(questions) = 'array' LOOP
        questions_arr := '[]'::jsonb;

        FOR q IN SELECT * FROM jsonb_array_elements(r.questions) LOOP
            new_q := q;

            IF new_q ? 'mcq' AND jsonb_typeof(new_q->'mcq') = 'object' THEN
                question_type := 'MCQ';
            ELSIF new_q ? 'tfq' AND jsonb_typeof(new_q->'tfq') = 'object' THEN
                question_type := 'TRUE_FALSE';
            ELSIF new_q ? 'frq' AND jsonb_typeof(new_q->'frq') = 'object' THEN
                question_type := 'FRQ';
            ELSE
                question_type := 'MCQ';
            END IF;

            IF NOT (new_q ? 'questionType') THEN
                new_q := jsonb_set(new_q, '{questionType}', to_jsonb(question_type));
            END IF;

            -- Rename tfq -> trueFalseQuestion
            IF new_q ? 'tfq' THEN
                new_q := jsonb_set(new_q, '{trueFalseQuestion}', new_q->'tfq');
                new_q := new_q - 'tfq';
            END IF;

            -- Reverse MCQ: answeredIndex -> answeredTerm
            IF new_q ? 'mcq' AND jsonb_typeof(new_q->'mcq') = 'object' AND new_q->'mcq' ? 'answeredIndex' THEN
                mcq_obj := new_q->'mcq';
                answered_idx := (mcq_obj->>'answeredIndex')::int;
                cci := (mcq_obj->>'correctChoiceIndex')::int;

                IF (mcq_obj->>'correct') = 'true' THEN
                    mcq_obj := jsonb_set(mcq_obj, '{answeredTerm}', mcq_obj->'term');
                ELSE
                    IF mcq_obj ? 'distractors' AND jsonb_typeof(mcq_obj->'distractors') = 'array' AND cci IS NOT NULL THEN
                        IF answered_idx < cci THEN
                            mcq_obj := jsonb_set(mcq_obj, '{answeredTerm}', COALESCE(mcq_obj->'distractors'->answered_idx, mcq_obj->'term'));
                        ELSIF answered_idx > cci THEN
                            mcq_obj := jsonb_set(mcq_obj, '{answeredTerm}', COALESCE(mcq_obj->'distractors'->(answered_idx - 1), mcq_obj->'term'));
                        ELSE
                            -- answered_idx == cci but correct is false? shouldn't happen but fallback to term
                            mcq_obj := jsonb_set(mcq_obj, '{answeredTerm}', mcq_obj->'term');
                        END IF;
                    ELSE
                        mcq_obj := jsonb_set(mcq_obj, '{answeredTerm}', mcq_obj->'term');
                        IF cci IS NULL THEN
                            mcq_obj := jsonb_set(mcq_obj, '{correctChoiceIndex}', to_jsonb(0));
                        END IF;
                    END IF;
                END IF;

                mcq_obj := mcq_obj - 'answeredIndex';
                new_q := jsonb_set(new_q, '{mcq}', mcq_obj);
            END IF;

            questions_arr := questions_arr || new_q;
        END LOOP;

        UPDATE practice_tests SET questions = questions_arr WHERE id = r.id;
    END LOOP;
END;
$$;
