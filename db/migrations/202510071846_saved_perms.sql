-- migrate:up
grant select on saved_studysets to quizfreely_api;
grant insert on saved_studysets to quizfreely_api;
grant update on saved_studysets to quizfreely_api;
grant delete on saved_studysets to quizfreely_api;

grant select on folders to quizfreely_api;
grant insert on folders to quizfreely_api;
grant update on folders to quizfreely_api;
grant delete on folders to quizfreely_api;

-- migrate:down
