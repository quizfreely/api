-- migrate:up
grant select on term_progress_history to quizfreely_api;
grant insert on term_progress_history to quizfreely_api;
grant delete on term_progress_history to quizfreely_api;

-- migrate:down
