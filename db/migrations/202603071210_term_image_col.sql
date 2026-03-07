-- migrate:up
create table images (
    object_key text primary key
);

alter table terms
add column term_image_key text references images(object_key),
add column def_image_key text references images(object_key);

create index idx_terms_term_image_key
on terms(term_image_key);

create index idx_terms_def_image_key
on terms(def_image_key);

insert into images (object_key)
select distinct object_key
from term_images
where object_key is not null;

update terms t
set term_image_key = ti.object_key
from term_images ti
where ti.term_id = t.id
and ti.def_side = false;

update terms t
set def_image_key = ti.object_key
from term_images ti
where ti.term_id = t.id
and ti.def_side = true;

drop table term_images;

-- migrate:down
create table term_images (
    object_key text primary key,
    def_side boolean not null,
    term_id uuid references terms (id) on delete set null
);

insert into term_images (term_id, object_key, def_side)
select id, term_image_key, false
from terms
where term_image_key is not null;

insert into term_images (term_id, object_key, def_side)
select id, def_image_key, true
from terms
where def_image_key is not null;

drop index if exists idx_terms_term_image_key;
drop index if exists idx_terms_def_image_key;

alter table terms
drop column term_image_key,
drop column def_image_key;

drop table images;
