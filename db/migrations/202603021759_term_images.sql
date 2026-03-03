-- migrate:up
create table term_images (
    object_key text primary key,
    def_side boolean not null,
    term_id uuid references terms (id) on delete set null
);

insert into term_images (object_key, def_side, term_id)
select term_image_key, false, id
from terms
where term_image_key is not null;

insert into term_images (object_key, def_side, term_id)
select def_image_key, true, id
from terms
where def_image_key is not null;

alter table terms
drop column term_image_key,
drop column def_image_key;

-- migrate:down
alter table terms 
add column term_image_key text,
add column def_image_key text;

update terms
set term_image_key = (
    select object_key
    from term_images
    where term_id = terms.id and def_side = false
    limit 1
)
where exists (
    select 1
    from term_images
    where term_id = terms.id and def_side = false
);

update terms
set def_image_key = (
    select object_key
    from term_images
    where term_id = terms.id and def_side = true
    limit 1
)
where exists (
    select 1
    from term_images
    where term_id = terms.id and def_side = true
);

drop table term_images;