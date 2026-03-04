-- migrate:up
ALTER TABLE term_images
ADD CONSTRAINT term_images_unique_term_id_and_def_side UNIQUE (term_id, def_side);

-- migrate:down
ALTER TABLE term_images
DROP CONSTRAINT term_images_unique_term_id_and_def_side;