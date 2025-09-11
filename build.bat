@echo off

:: Zielordner sicherstellen
if not exist ..\dist mkdir ..\dist
if not exist ..\dist\assets mkdir ..\dist\assets

:: Builds
esbuild frontend/assets/js/usermanager.js    --bundle --minify --format=esm --target=es2022 --outfile=js/usermanager.js
esbuild frontend/assets/js/routemanager.js --bundle --minify --format=esm --target=es2022 --outfile=js/routemanager.js
esbuild frontend/assets/js/index.js    --bundle --minify --format=esm --target=es2022 --outfile=js/index.js

:: HTML kopieren (überschreiben)
cd frontend
copy /Y index.html  ..\dist\
copy /Y routes.html ..\dist\
copy /Y users.html  ..\dist\

:: JS nach assets kopieren
robocopy js ..\dist\assets /E
robocopy assets\style ..\dist\assets /E


DB=speedliner
docker exec -it 7044c01a28a3 psql -U speedliner -d "$DB" -c \
"DROP SCHEMA public CASCADE; CREATE SCHEMA public; ALTER SCHEMA public OWNER TO speedliner; GRANT ALL ON SCHEMA public TO speedliner; GRANT ALL ON SCHEMA public TO public;"
