-- migrate:up
alter table terms 
add column term_image_key text,
add column def_image_key text;
-- migrate:down
alter table terms 
drop column term_image_key,
drop column def_image_key;
