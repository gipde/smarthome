# Welcome to smarthome

## Smarthome ideen
- Rolladen
- Türkontakte
- Fensterkontakte
- Gassensor
- Temperatursensor
    - Raum
    - Heizung
- Feuchtesensor
- Garagentor
- IP Kamera

## Passwort-Hash
Passwort Hashes werden für verschiedene Dinge benötigt
- OAuth2
- Passwortablage der User

Um einen gültigen Passwort-Hash zu erhalten bitte den folgenden Controller aufrufen
http://localhost:9000/main/GetHash?password=foobar


## Cross-compilation

- Linux:  APP_VERSION=0.1.0  GOOS=linux GOARCH=amd64 revel package schneidernet/smarthome
- RPi:    APP_VERSION=0.1.0  GOOS=linux GOARCH=arm GOARM=6 revel package schneidernet/smarthome

## TODO:
