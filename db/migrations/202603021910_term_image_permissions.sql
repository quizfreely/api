-- migrate:up
grant select on term_images to quizfreely_api;
grant insert on term_images to quizfreely_api;
grant update on term_images to quizfreely_api;
grant delete on term_images to quizfreely_api;

-- migrate:down
revoke select on term_images to quizfreely_api;
revoke insert on term_images to quizfreely_api;
revoke update on term_images to quizfreely_api;
revoke delete on term_images to quizfreely_api;