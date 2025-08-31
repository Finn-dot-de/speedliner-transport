--
-- PostgreSQL database cluster dump
--

SET default_transaction_read_only = off;

SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;

--
-- Roles
--

CREATE ROLE speedliner;
ALTER ROLE speedliner WITH SUPERUSER INHERIT CREATEROLE CREATEDB LOGIN REPLICATION BYPASSRLS PASSWORD 'SCRAM-SHA-256$4096:A3LGFkdciKBIpN1Syn4V3g==$HC55naCbeBDH1eHFcRmo1EQRE7zM70j3rE8S2Ya5EWU=:lbYlOy9d/WY9S3rOTicfmrW5exw70Pmy9OGDlyNLEpE=';

--
-- User Configurations
--








--
-- Databases
--

--
-- Database "template1" dump
--

\connect template1

--
-- PostgreSQL database dump
--

-- Dumped from database version 15.10 (Debian 15.10-1.pgdg120+1)
-- Dumped by pg_dump version 15.10 (Debian 15.10-1.pgdg120+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- PostgreSQL database dump complete
--

--
-- Database "postgres" dump
--

\connect postgres

--
-- PostgreSQL database dump
--

-- Dumped from database version 15.10 (Debian 15.10-1.pgdg120+1)
-- Dumped by pg_dump version 15.10 (Debian 15.10-1.pgdg120+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- PostgreSQL database dump complete
--

--
-- Database "speedliner" dump
--

--
-- PostgreSQL database dump
--

-- Dumped from database version 15.10 (Debian 15.10-1.pgdg120+1)
-- Dumped by pg_dump version 15.10 (Debian 15.10-1.pgdg120+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: speedliner; Type: DATABASE; Schema: -; Owner: speedliner
--

CREATE DATABASE speedliner WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'en_US.utf8';


ALTER DATABASE speedliner OWNER TO speedliner;

\connect speedliner

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: routes; Type: TABLE; Schema: public; Owner: speedliner
--

CREATE TABLE public.routes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    from_system text NOT NULL,
    to_system text NOT NULL,
    price_per_m3 numeric(10,2)
);


ALTER TABLE public.routes OWNER TO speedliner;

--
-- Name: users; Type: TABLE; Schema: public; Owner: speedliner
--

CREATE TABLE public.users (
    char_id bigint NOT NULL,
    name text NOT NULL,
    role text DEFAULT 'user'::text NOT NULL
);


ALTER TABLE public.users OWNER TO speedliner;

--
-- Data for Name: routes; Type: TABLE DATA; Schema: public; Owner: speedliner
--

COPY public.routes (id, from_system, to_system, price_per_m3) FROM stdin;
6acaa281-8955-41c1-bf4d-c100d1173579	Jita	K-6K16	1050.00
90a592f3-49fc-44fc-8d5c-4a8f28763c13	B-9C24	K-6K16	1050.00
c02eb18b-9f87-415c-9cdf-f9bdafaf00a6	Amarr	K-6K16	900.00
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: speedliner
--

COPY public.users (char_id, name, role) FROM stdin;
2123452374	Shirok Daasek	user
92393462	Philippe Rochard	provider
2119669460	sMilaf	user
923693091	Korexx	user
2123597632	Event Horizon Sun	user
2123272217	Rusty Weld	user
94237906	Kyle Shaile	user
2118431553	Comander-Video	admin
\.


--
-- Name: routes routes_pkey; Type: CONSTRAINT; Schema: public; Owner: speedliner
--

ALTER TABLE ONLY public.routes
    ADD CONSTRAINT routes_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: speedliner
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (char_id);


--
-- PostgreSQL database dump complete
--

--
-- PostgreSQL database cluster dump complete
--

