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


## Update Go Toolchain
### Update Packages
go get -u all

### Update Revel
go get -u github.com/revel/cmd/revel

## new Go Installation
??? 

## Travis-CI



## Race Detector
go build schneidernet/smarthome target
modify target/run.sh
```
go run -race $SCRIPTPATH/src/schneidernet/smarthome/app/tmp/main.go  -importPath schneidernet/smarthome -srcPath "$SCRIPTPATH/src" -runMode dev
#"$SCRIPTPATH/smarthome" -importPath schneidernet/smarthome -srcPath "$SCRIPTPATH/src" -runMode dev
```

## TODO:
- FR: Device Passwort setzen
- FR: Purge Device-Logs
- FR: Versionierung der Anwendung und der kleinere Pakete wie rpiswitch
- FR: Persistierung von Sensordaten (z.B. Temperatur)
- FR: Visualisierung der historischen Sensordaten
- BUG: Beim Speichern in den Device-Settings (falscher Redirect)