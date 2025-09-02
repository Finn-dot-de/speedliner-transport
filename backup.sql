INSERT INTO routes (id, from_system, to_system, price_per_m3) VALUES
                                                                  ('6acaa281-8955-41c1-bf4d-c100d1173579','Jita','K-6K16',1050.00),
                                                                  ('90a592f3-49fc-44fc-8d5c-4a8f28763c13','B-9C24','K-6K16',1050.00),
                                                                  ('c02eb18b-9f87-415c-9cdf-f9bdafaf00a6','Amarr','K-6K16',900.00)
ON CONFLICT (id) DO NOTHING;

INSERT INTO users (char_id, name, role) VALUES
                                            (2123452374,'Shirok Daasek','user'),
                                            (92393462,'Philippe Rochard','provider'),
                                            (2119669460,'sMilaf','user'),
                                            (923693091,'Korexx','user'),
                                            (2123597632,'Event Horizon Sun','user'),
                                            (2123272217,'Rusty Weld','user'),
                                            (94237906,'Kyle Shaile','user'),
                                            (2118431553,'Comander-Video','admin')
ON CONFLICT (char_id) DO NOTHING;
