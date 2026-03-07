-- migrate:up
grant select on images to quizfreely_api;
grant insert on images to quizfreely_api;
grant update on images to quizfreely_api;
grant delete on images to quizfreely_api;

-- migrate:down
revoke select on images to quizfreely_api;
revoke insert on images to quizfreely_api;
revoke update on images to quizfreely_api;
revoke delete on images to quizfreely_api;
