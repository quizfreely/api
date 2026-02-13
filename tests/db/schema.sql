SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: auth; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA auth;


--
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: answer_with_enum; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.answer_with_enum AS ENUM (
    'TERM',
    'DEF'
);


--
-- Name: auth_type_enum; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.auth_type_enum AS ENUM (
    'USERNAME_PASSWORD',
    'OAUTH_GOOGLE'
);


--
-- Name: subject_category; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.subject_category AS ENUM (
    'LANG',
    'STEM',
    'SOCIAL_STUDIES',
    'LA',
    'MATH'
);


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: sessions; Type: TABLE; Schema: auth; Owner: -
--

CREATE TABLE auth.sessions (
    id bigint NOT NULL,
    token text DEFAULT encode(public.gen_random_bytes(32), 'base64'::text) NOT NULL,
    user_id uuid NOT NULL,
    expire_at timestamp with time zone DEFAULT (now() + '10 days'::interval)
);


--
-- Name: sessions_id_seq; Type: SEQUENCE; Schema: auth; Owner: -
--

CREATE SEQUENCE auth.sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: auth; Owner: -
--

ALTER SEQUENCE auth.sessions_id_seq OWNED BY auth.sessions.id;


--
-- Name: users; Type: TABLE; Schema: auth; Owner: -
--

CREATE TABLE auth.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    username text,
    encrypted_password text,
    display_name text NOT NULL,
    auth_type public.auth_type_enum NOT NULL,
    oauth_google_sub text,
    oauth_google_email text,
    mod_perms boolean DEFAULT false NOT NULL,
    oauth_google_name text
);


--
-- Name: folder_studysets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.folder_studysets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    studyset_id uuid NOT NULL,
    user_id uuid NOT NULL,
    folder_id uuid NOT NULL,
    "timestamp" timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: folders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.folders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    name text NOT NULL
);


--
-- Name: practice_tests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.practice_tests (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    "timestamp" timestamp with time zone DEFAULT now() NOT NULL,
    user_id uuid NOT NULL,
    studyset_id uuid NOT NULL,
    questions_correct smallint,
    questions_total smallint,
    questions jsonb
);


--
-- Name: saved_studysets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saved_studysets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    studyset_id uuid,
    user_id uuid,
    "timestamp" timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying NOT NULL
);


--
-- Name: studysets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.studysets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    title text DEFAULT 'Untitled Studyset'::text NOT NULL,
    private boolean DEFAULT false NOT NULL,
    updated_at timestamp with time zone DEFAULT now(),
    terms_count integer,
    tsvector_title tsvector GENERATED ALWAYS AS (to_tsvector('english'::regconfig, title)) STORED,
    subject_id text,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: subject_keywords; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.subject_keywords (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    keyword text,
    subject_id text
);


--
-- Name: subjects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.subjects (
    id text NOT NULL,
    name text,
    category public.subject_category
);


--
-- Name: term_confusion_pairs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.term_confusion_pairs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    term_id uuid NOT NULL,
    confused_term_id uuid NOT NULL,
    answered_with public.answer_with_enum NOT NULL,
    confused_count integer,
    last_confused_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: term_progress; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.term_progress (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    term_id uuid NOT NULL,
    user_id uuid NOT NULL,
    term_first_reviewed_at timestamp with time zone,
    term_last_reviewed_at timestamp with time zone,
    term_review_count integer,
    def_first_reviewed_at timestamp with time zone,
    def_last_reviewed_at timestamp with time zone,
    def_review_count integer,
    term_leitner_system_box smallint,
    def_leitner_system_box smallint,
    term_correct_count integer DEFAULT 0 NOT NULL,
    term_incorrect_count integer DEFAULT 0 NOT NULL,
    def_correct_count integer DEFAULT 0 NOT NULL,
    def_incorrect_count integer DEFAULT 0 NOT NULL
);


--
-- Name: term_progress_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.term_progress_history (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    "timestamp" timestamp with time zone DEFAULT now() NOT NULL,
    term_id uuid NOT NULL,
    user_id uuid NOT NULL,
    term_correct_count integer,
    term_incorrect_count integer,
    def_correct_count integer,
    def_incorrect_count integer
);


--
-- Name: terms; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.terms (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    term text,
    def text,
    studyset_id uuid NOT NULL,
    sort_order integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: sessions id; Type: DEFAULT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.sessions ALTER COLUMN id SET DEFAULT nextval('auth.sessions_id_seq'::regclass);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: users users_oauth_google_sub_key; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.users
    ADD CONSTRAINT users_oauth_google_sub_key UNIQUE (oauth_google_sub);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: term_confusion_pairs confusion_pairs_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_confusion_pairs
    ADD CONSTRAINT confusion_pairs_unique UNIQUE (user_id, term_id, confused_term_id, answered_with);


--
-- Name: folder_studysets folder_studysets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.folder_studysets
    ADD CONSTRAINT folder_studysets_pkey PRIMARY KEY (id);


--
-- Name: folder_studysets folder_studysets_studyset_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.folder_studysets
    ADD CONSTRAINT folder_studysets_studyset_id_user_id_key UNIQUE (studyset_id, user_id);


--
-- Name: folders folders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.folders
    ADD CONSTRAINT folders_pkey PRIMARY KEY (id);


--
-- Name: practice_tests practice_tests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.practice_tests
    ADD CONSTRAINT practice_tests_pkey PRIMARY KEY (id);


--
-- Name: saved_studysets saved_studysets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_studysets
    ADD CONSTRAINT saved_studysets_pkey PRIMARY KEY (id);


--
-- Name: saved_studysets saved_studysets_user_studyset_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_studysets
    ADD CONSTRAINT saved_studysets_user_studyset_unique UNIQUE (user_id, studyset_id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: studysets studysets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.studysets
    ADD CONSTRAINT studysets_pkey PRIMARY KEY (id);


--
-- Name: subject_keywords subject_keywords_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subject_keywords
    ADD CONSTRAINT subject_keywords_pkey PRIMARY KEY (id);


--
-- Name: subjects subjects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subjects
    ADD CONSTRAINT subjects_pkey PRIMARY KEY (id);


--
-- Name: term_confusion_pairs term_confusion_pairs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_confusion_pairs
    ADD CONSTRAINT term_confusion_pairs_pkey PRIMARY KEY (id);


--
-- Name: term_progress_history term_progress_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_progress_history
    ADD CONSTRAINT term_progress_history_pkey PRIMARY KEY (id);


--
-- Name: term_progress term_progress_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_progress
    ADD CONSTRAINT term_progress_pkey PRIMARY KEY (id);


--
-- Name: term_progress term_progress_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_progress
    ADD CONSTRAINT term_progress_unique UNIQUE (term_id, user_id);


--
-- Name: terms terms_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.terms
    ADD CONSTRAINT terms_pkey PRIMARY KEY (id);


--
-- Name: studysets_title_trgm_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX studysets_title_trgm_idx ON public.studysets USING gin (lower(title) public.gin_trgm_ops);


--
-- Name: subject_keywords_trgm_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX subject_keywords_trgm_idx ON public.subject_keywords USING gin (keyword public.gin_trgm_ops);


--
-- Name: textsearch_title_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX textsearch_title_idx ON public.studysets USING gin (tsvector_title);


--
-- Name: sessions sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.sessions
    ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;


--
-- Name: folder_studysets folder_studysets_folder_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.folder_studysets
    ADD CONSTRAINT folder_studysets_folder_id_fkey FOREIGN KEY (folder_id) REFERENCES public.folders(id) ON DELETE CASCADE;


--
-- Name: folder_studysets folder_studysets_studyset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.folder_studysets
    ADD CONSTRAINT folder_studysets_studyset_id_fkey FOREIGN KEY (studyset_id) REFERENCES public.studysets(id) ON DELETE CASCADE;


--
-- Name: folder_studysets folder_studysets_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.folder_studysets
    ADD CONSTRAINT folder_studysets_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;


--
-- Name: folders folders_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.folders
    ADD CONSTRAINT folders_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;


--
-- Name: practice_tests practice_tests_studyset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.practice_tests
    ADD CONSTRAINT practice_tests_studyset_id_fkey FOREIGN KEY (studyset_id) REFERENCES public.studysets(id) ON DELETE CASCADE;


--
-- Name: practice_tests practice_tests_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.practice_tests
    ADD CONSTRAINT practice_tests_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;


--
-- Name: saved_studysets saved_studysets_studyset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_studysets
    ADD CONSTRAINT saved_studysets_studyset_id_fkey FOREIGN KEY (studyset_id) REFERENCES public.studysets(id) ON DELETE CASCADE;


--
-- Name: saved_studysets saved_studysets_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saved_studysets
    ADD CONSTRAINT saved_studysets_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;


--
-- Name: studysets studysets_subject_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.studysets
    ADD CONSTRAINT studysets_subject_id_fkey FOREIGN KEY (subject_id) REFERENCES public.subjects(id) ON DELETE SET NULL;


--
-- Name: studysets studysets_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.studysets
    ADD CONSTRAINT studysets_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE SET NULL;


--
-- Name: subject_keywords subject_keywords_subject_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subject_keywords
    ADD CONSTRAINT subject_keywords_subject_id_fkey FOREIGN KEY (subject_id) REFERENCES public.subjects(id) ON DELETE CASCADE;


--
-- Name: term_confusion_pairs term_confusion_pairs_confused_term_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_confusion_pairs
    ADD CONSTRAINT term_confusion_pairs_confused_term_id_fkey FOREIGN KEY (confused_term_id) REFERENCES public.terms(id) ON DELETE CASCADE;


--
-- Name: term_confusion_pairs term_confusion_pairs_term_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_confusion_pairs
    ADD CONSTRAINT term_confusion_pairs_term_id_fkey FOREIGN KEY (term_id) REFERENCES public.terms(id) ON DELETE CASCADE;


--
-- Name: term_confusion_pairs term_confusion_pairs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_confusion_pairs
    ADD CONSTRAINT term_confusion_pairs_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;


--
-- Name: term_progress_history term_progress_history_term_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_progress_history
    ADD CONSTRAINT term_progress_history_term_id_fkey FOREIGN KEY (term_id) REFERENCES public.terms(id) ON DELETE CASCADE;


--
-- Name: term_progress_history term_progress_history_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_progress_history
    ADD CONSTRAINT term_progress_history_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;


--
-- Name: term_progress term_progress_term_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_progress
    ADD CONSTRAINT term_progress_term_id_fkey FOREIGN KEY (term_id) REFERENCES public.terms(id) ON DELETE CASCADE;


--
-- Name: term_progress term_progress_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_progress
    ADD CONSTRAINT term_progress_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;


--
-- Name: terms terms_studyset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.terms
    ADD CONSTRAINT terms_studyset_id_fkey FOREIGN KEY (studyset_id) REFERENCES public.studysets(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('202508140123'),
    ('202508141431'),
    ('202508181513'),
    ('202508191404'),
    ('202508201847'),
    ('202508202155'),
    ('202508211445'),
    ('202509021427'),
    ('202509030947'),
    ('202509302346'),
    ('202510010013'),
    ('202510021734'),
    ('202510021832'),
    ('202510061818'),
    ('202510061833'),
    ('202510062336'),
    ('202510071846'),
    ('202510091706'),
    ('202510122101'),
    ('202510292200'),
    ('202602101650');
