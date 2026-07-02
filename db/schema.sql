\restrict BNHcv39vDkDMzy4uZuRsGdZmnpTOnueAVQigO9ggzBixUe9qdtCldOJsARZhK3t

-- Dumped from database version 18.4
-- Dumped by pg_dump version 18.4

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
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

-- *not* creating schema, since initdb creates it


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
-- Name: auth_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.auth_type AS ENUM (
    'USERNAME_PASSWORD',
    'OAUTH_GOOGLE'
);


--
-- Name: fsrs_rating; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.fsrs_rating AS ENUM (
    'MANUAL',
    'AGAIN',
    'HARD',
    'GOOD',
    'EASY'
);


--
-- Name: fsrs_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.fsrs_state AS ENUM (
    'NEW',
    'LEARNING',
    'REVIEW',
    'RELEARNING'
);


--
-- Name: question_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.question_type AS ENUM (
    'MCQ',
    'TFQ',
    'FRQ'
);


--
-- Name: review_activity_type_enum; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.review_activity_type_enum AS ENUM (
    'PRACTICE_TEST',
    'MATCH'
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
    token text DEFAULT encode(public.gen_random_bytes(32), 'base64'::text) NOT NULL,
    user_id uuid NOT NULL,
    expire_at timestamp with time zone DEFAULT (now() + '10 days'::interval)
);


--
-- Name: users; Type: TABLE; Schema: auth; Owner: -
--

CREATE TABLE auth.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    username text,
    encrypted_password text,
    display_name text NOT NULL,
    auth_type public.auth_type NOT NULL,
    oauth_google_sub text,
    oauth_google_email text,
    mod_perms boolean DEFAULT false NOT NULL,
    oauth_google_name text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
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
    name text NOT NULL,
    private boolean DEFAULT true NOT NULL
);


--
-- Name: fsrs_cards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.fsrs_cards (
    term_id uuid NOT NULL,
    user_id uuid NOT NULL,
    difficulty double precision NOT NULL,
    due timestamp with time zone NOT NULL,
    lapses integer NOT NULL,
    last_review timestamp with time zone,
    learning_steps integer NOT NULL,
    reps integer NOT NULL,
    scheduled_days integer NOT NULL,
    stability double precision NOT NULL,
    state public.fsrs_state NOT NULL
);


--
-- Name: fsrs_review_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.fsrs_review_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    term_id uuid,
    user_id uuid,
    difficulty double precision NOT NULL,
    due timestamp with time zone NOT NULL,
    learning_steps integer NOT NULL,
    rating public.fsrs_rating NOT NULL,
    review timestamp with time zone NOT NULL,
    scheduled_days integer NOT NULL,
    stability double precision NOT NULL,
    state public.fsrs_state NOT NULL
);


--
-- Name: images; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.images (
    object_key text NOT NULL
);


--
-- Name: match_activities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.match_activities (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    duration_ms integer NOT NULL,
    end_timestamp timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: match_activity_studysets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.match_activity_studysets (
    match_id uuid NOT NULL,
    studyset_id uuid NOT NULL
);


--
-- Name: practice_test_questions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.practice_test_questions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    practice_test_id uuid NOT NULL,
    term_id uuid,
    term_snapshot text NOT NULL,
    def_snapshot text NOT NULL,
    type public.question_type NOT NULL,
    answer_with public.answer_with_enum NOT NULL,
    correct boolean NOT NULL,
    "position" integer NOT NULL,
    data jsonb NOT NULL
);


--
-- Name: practice_test_studysets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.practice_test_studysets (
    practice_test_id uuid NOT NULL,
    studyset_id uuid NOT NULL
);


--
-- Name: practice_tests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.practice_tests (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    "timestamp" timestamp with time zone DEFAULT now() NOT NULL,
    user_id uuid NOT NULL,
    questions_correct smallint,
    questions_total smallint
);


--
-- Name: review_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.review_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    term_id uuid NOT NULL,
    practice_test_question_id uuid,
    correct boolean NOT NULL,
    answer_with public.answer_with_enum,
    "timestamp" timestamp with time zone DEFAULT now() NOT NULL,
    answered_term_id uuid,
    practice_test_question_type public.question_type,
    review_activity_type public.review_activity_type_enum NOT NULL,
    answered_string text,
    match_activity_id uuid
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
    title text NOT NULL,
    private boolean NOT NULL,
    updated_at timestamp with time zone DEFAULT now(),
    terms_count integer,
    tsvector_title tsvector GENERATED ALWAYS AS (to_tsvector('english'::regconfig, title)) STORED,
    subject_id text,
    created_at timestamp with time zone DEFAULT now(),
    draft boolean DEFAULT false NOT NULL,
    seo_indexing_approved boolean DEFAULT false NOT NULL
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
    term_correct_count integer DEFAULT 0 NOT NULL,
    term_incorrect_count integer DEFAULT 0 NOT NULL,
    def_correct_count integer DEFAULT 0 NOT NULL,
    def_incorrect_count integer DEFAULT 0 NOT NULL
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
    updated_at timestamp with time zone DEFAULT now(),
    term_image_key text,
    def_image_key text
);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (token);


--
-- Name: users users_oauth_google_id_key; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.users
    ADD CONSTRAINT users_oauth_google_id_key UNIQUE (oauth_google_sub);


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
-- Name: fsrs_cards fsrs_cards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fsrs_cards
    ADD CONSTRAINT fsrs_cards_pkey PRIMARY KEY (term_id, user_id);


--
-- Name: fsrs_review_logs fsrs_review_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fsrs_review_logs
    ADD CONSTRAINT fsrs_review_logs_pkey PRIMARY KEY (id);


--
-- Name: images images_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.images
    ADD CONSTRAINT images_pkey PRIMARY KEY (object_key);


--
-- Name: match_activities match_activities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.match_activities
    ADD CONSTRAINT match_activities_pkey PRIMARY KEY (id);


--
-- Name: match_activity_studysets match_activity_studysets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.match_activity_studysets
    ADD CONSTRAINT match_activity_studysets_pkey PRIMARY KEY (match_id, studyset_id);


--
-- Name: practice_test_questions practice_test_questions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.practice_test_questions
    ADD CONSTRAINT practice_test_questions_pkey PRIMARY KEY (id);


--
-- Name: practice_test_questions practice_test_questions_practice_test_id_position_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.practice_test_questions
    ADD CONSTRAINT practice_test_questions_practice_test_id_position_key UNIQUE (practice_test_id, "position");


--
-- Name: practice_test_studysets practice_test_studysets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.practice_test_studysets
    ADD CONSTRAINT practice_test_studysets_pkey PRIMARY KEY (practice_test_id, studyset_id);


--
-- Name: practice_tests practice_tests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.practice_tests
    ADD CONSTRAINT practice_tests_pkey PRIMARY KEY (id);


--
-- Name: review_events review_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.review_events
    ADD CONSTRAINT review_events_pkey PRIMARY KEY (id);


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
-- Name: idx_pts_studyset_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_pts_studyset_id ON public.practice_test_studysets USING btree (studyset_id);


--
-- Name: idx_review_events_practice_test_question_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_review_events_practice_test_question_id ON public.review_events USING btree (practice_test_question_id);


--
-- Name: idx_review_events_term_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_review_events_term_id ON public.review_events USING btree (term_id);


--
-- Name: idx_review_events_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_review_events_user_id ON public.review_events USING btree (user_id);


--
-- Name: idx_terms_def_image_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_terms_def_image_key ON public.terms USING btree (def_image_key);


--
-- Name: idx_terms_term_image_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_terms_term_image_key ON public.terms USING btree (term_image_key);


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
-- Name: fsrs_cards fsrs_cards_term_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fsrs_cards
    ADD CONSTRAINT fsrs_cards_term_id_fkey FOREIGN KEY (term_id) REFERENCES public.terms(id);


--
-- Name: fsrs_cards fsrs_cards_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fsrs_cards
    ADD CONSTRAINT fsrs_cards_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id);


--
-- Name: fsrs_review_logs fsrs_review_logs_term_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fsrs_review_logs
    ADD CONSTRAINT fsrs_review_logs_term_id_fkey FOREIGN KEY (term_id) REFERENCES public.terms(id);


--
-- Name: fsrs_review_logs fsrs_review_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fsrs_review_logs
    ADD CONSTRAINT fsrs_review_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id);


--
-- Name: match_activities match_activities_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.match_activities
    ADD CONSTRAINT match_activities_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;


--
-- Name: match_activity_studysets match_activity_studysets_match_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.match_activity_studysets
    ADD CONSTRAINT match_activity_studysets_match_id_fkey FOREIGN KEY (match_id) REFERENCES public.match_activities(id) ON DELETE CASCADE;


--
-- Name: match_activity_studysets match_activity_studysets_studyset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.match_activity_studysets
    ADD CONSTRAINT match_activity_studysets_studyset_id_fkey FOREIGN KEY (studyset_id) REFERENCES public.studysets(id) ON DELETE CASCADE;


--
-- Name: practice_test_questions practice_test_questions_practice_test_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.practice_test_questions
    ADD CONSTRAINT practice_test_questions_practice_test_id_fkey FOREIGN KEY (practice_test_id) REFERENCES public.practice_tests(id) ON DELETE CASCADE;


--
-- Name: practice_test_questions practice_test_questions_term_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.practice_test_questions
    ADD CONSTRAINT practice_test_questions_term_id_fkey FOREIGN KEY (term_id) REFERENCES public.terms(id) ON DELETE SET NULL;


--
-- Name: practice_test_studysets practice_test_studysets_practice_test_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.practice_test_studysets
    ADD CONSTRAINT practice_test_studysets_practice_test_id_fkey FOREIGN KEY (practice_test_id) REFERENCES public.practice_tests(id) ON DELETE CASCADE;


--
-- Name: practice_test_studysets practice_test_studysets_studyset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.practice_test_studysets
    ADD CONSTRAINT practice_test_studysets_studyset_id_fkey FOREIGN KEY (studyset_id) REFERENCES public.studysets(id) ON DELETE CASCADE;


--
-- Name: practice_tests practice_tests_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.practice_tests
    ADD CONSTRAINT practice_tests_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;


--
-- Name: review_events review_events_answered_term_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.review_events
    ADD CONSTRAINT review_events_answered_term_id_fkey FOREIGN KEY (answered_term_id) REFERENCES public.terms(id) ON DELETE CASCADE;


--
-- Name: review_events review_events_match_activity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.review_events
    ADD CONSTRAINT review_events_match_activity_id_fkey FOREIGN KEY (match_activity_id) REFERENCES public.match_activities(id) ON DELETE CASCADE;


--
-- Name: review_events review_events_practice_test_question_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.review_events
    ADD CONSTRAINT review_events_practice_test_question_id_fkey FOREIGN KEY (practice_test_question_id) REFERENCES public.practice_test_questions(id) ON DELETE CASCADE;


--
-- Name: review_events review_events_term_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.review_events
    ADD CONSTRAINT review_events_term_id_fkey FOREIGN KEY (term_id) REFERENCES public.terms(id) ON DELETE CASCADE;


--
-- Name: review_events review_events_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.review_events
    ADD CONSTRAINT review_events_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;


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
-- Name: terms terms_def_image_key_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.terms
    ADD CONSTRAINT terms_def_image_key_fkey FOREIGN KEY (def_image_key) REFERENCES public.images(object_key);


--
-- Name: terms terms_studyset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.terms
    ADD CONSTRAINT terms_studyset_id_fkey FOREIGN KEY (studyset_id) REFERENCES public.studysets(id) ON DELETE CASCADE;


--
-- Name: terms terms_term_image_key_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.terms
    ADD CONSTRAINT terms_term_image_key_fkey FOREIGN KEY (term_image_key) REFERENCES public.images(object_key);


--
-- PostgreSQL database dump complete
--

\unrestrict BNHcv39vDkDMzy4uZuRsGdZmnpTOnueAVQigO9ggzBixUe9qdtCldOJsARZhK3t


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
    ('202602101650'),
    ('202602241455'),
    ('202602281700'),
    ('202603021759'),
    ('202603021910'),
    ('202603031915'),
    ('202603071210'),
    ('202603071640'),
    ('202603212200'),
    ('202605031122'),
    ('202605031337'),
    ('202606170845'),
    ('202606180946'),
    ('202606251140'),
    ('202606252121'),
    ('202606261000'),
    ('202606290020'),
    ('202606291000'),
    ('202606291100'),
    ('202606301100'),
    ('202607012025');
