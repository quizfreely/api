-- migrate:up
drop table if exists fsrs_cards;
drop table if exists fsrs_review_logs;

create table fsrs_cards (
    term_id uuid not null references terms(id) on delete cascade,
    user_id uuid not null references auth.users(id) on delete cascade,
    difficulty real not null,
    due timestamptz not null,
    lapses int not null,
    last_review timestamptz,
    learning_steps int not null,
    reps int not null,
    scheduled_days int not null,
    stability real not null,
    state fsrs_state not null,
    primary key (term_id, user_id)
);
grant select, insert, update, delete on fsrs_cards to quizfreely_api;

create table fsrs_review_logs (
    id uuid primary key default gen_random_uuid(),
    term_id uuid not null references terms(id) on delete cascade,
    user_id uuid not null references auth.users(id) on delete cascade,
    difficulty real not null,
    due timestamptz not null,
    learning_steps int not null,
    rating fsrs_rating not null,
    review timestamptz not null,
    scheduled_days int not null,
    stability real not null,
    state fsrs_state not null
);
grant select, insert, update, delete on fsrs_review_logs to quizfreely_api;

-- migrate:down
drop table if exists fsrs_cards;
drop table if exists fsrs_review_logs;

create table fsrs_cards (
    term_id uuid references terms(id),
    user_id uuid references auth.users(id),
    difficulty double precision not null,
    due timestamptz not null,
    lapses int not null,
    last_review timestamptz,
    learning_steps int not null,
    reps int not null,
    scheduled_days int not null,
    stability double precision not null,
    state fsrs_state not null,
    primary key (term_id, user_id)
);
grant select, insert, update, delete on fsrs_cards to quizfreely_api;

create table fsrs_review_logs (
    id uuid primary key default gen_random_uuid(),
    term_id uuid references terms(id),
    user_id uuid references auth.users(id),
    difficulty double precision not null,
    due timestamptz not null,
    learning_steps int not null,
    rating fsrs_rating not null,
    review timestamptz not null,
    scheduled_days int not null,
    stability double precision not null,
    state fsrs_state not null
);
grant select, insert, update, delete on fsrs_review_logs to quizfreely_api;

