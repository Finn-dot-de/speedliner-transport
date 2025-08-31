-- portable_single_db.sql
\set ON_ERROR_STOP on

-- Saubere Session-Einstellungen
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

-- Immer ins public-Schema arbeiten
CREATE SCHEMA IF NOT EXISTS public AUTHORIZATION CURRENT_USER;
SET search_path = public;

-- UUID-Generator (für gen_random_uuid)
CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;

-- Tabellen (idempotent)
CREATE TABLE IF NOT EXISTS public.routes (
                                             id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    from_system  text NOT NULL,
    to_system    text NOT NULL,
    price_per_m3 numeric(10,2)
    );

CREATE TABLE IF NOT EXISTS public.users (
                                            char_id bigint PRIMARY KEY,
                                            name    text NOT NULL,
                                            role    text NOT NULL DEFAULT 'user'
);

-- Optional: Ownership auf aktuellen User setzen (robust, kein fester Name)
ALTER TABLE public.routes OWNER TO CURRENT_USER;
ALTER TABLE public.users  OWNER TO CURRENT_USER;

-- Daten laden (löscht nichts; fügt nur hinzu, überspringt Duplikate)
-- Nutzt INSERT ... ON CONFLICT statt COPY -> besser portabel & wiederholbar
INSERT INTO public.routes (id, from_system, to_system, price_per_m3) VALUES
                                                                         ('6acaa281-8955-41c1-bf4d-c100d1173579','Jita','K-6K16',1050.00),
                                                                         ('90a592f3-49fc-44fc-8d5c-4a8f28763c13','B-9C24','K-6K16',1050.00),
                                                                         ('c02eb18b-9f87-415c-9cdf-f9bdafaf00a6','Amarr','K-6K16',900.00)
    ON CONFLICT (id) DO NOTHING;

INSERT INTO public.users (char_id, name, role) VALUES
                                                   (2123452374,'Shirok Daasek','user'),
                                                   (92393462,'Philippe Rochard','provider'),
                                                   (2119669460,'sMilaf','user'),
                                                   (923693091,'Korexx','user'),
                                                   (2123597632,'Event Horizon Sun','user'),
                                                   (2123272217,'Rusty Weld','user'),
                                                   (94237906,'Kyle Shaile','user'),
                                                   (2118431553,'Comander-Video','admin')
    ON CONFLICT (char_id) DO NOTHING;
