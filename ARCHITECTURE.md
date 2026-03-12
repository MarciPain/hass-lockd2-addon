# Projekt Architektúra - hass-lockd2-addon

## Aktuális állapot
A projekt egy Home Assistant Addon, amely MQTT-n keresztül vezérel eszközöket. Legutóbb teljes mértékben implementálva lett az ESPHome-hoz hasonló MQTT publikálási logika, WebSocket kapcsolat a HA `state_changed` eseményeinek figyelésére, valamint kibővült az `app.js` entitásszűrője (akkumulátor/elem nevek kezelésével).

## Fejlesztési Irányok és Szabályok
- **SZABÁLY:** Minden kódmódosítás után, még a Git push (vagy a tesztelés/felhasználónak való átadás) előtt a `lockd2/config.yaml`-ben **kötelező növelni a verziószámot**. Ennek hiányában az Addon nem fog frissülni a Home Assistantban.
- MQTT kapcsolat stabilitásának biztosítása.
- Home Assistant entitás felfedezés (Discovery) finomhangolása.

## Fájllista és funkciók
- [lockd2/backend/api.go](./lockd2/backend/api.go): API végpontok és Ingress middleware.
- [lockd2/backend/ha_api.go](./lockd2/backend/ha_api.go): Supervisor API integráció (illetve CallHAService implementáció zárak/kapcsolók vezérlésére) és hitelesítés.
- [lockd2/backend/ha_ws.go](./lockd2/backend/ha_ws.go): ÚJ: Home Assistant WebSocket kliens az entitás `state_changed` eseményeinek valós idejű figyeléséhez.
- [lockd2/backend/mqtt.go](./lockd2/backend/mqtt.go): MQTT kliens MQTT publikálással és parancs (`cmd`) fogadással (Zár/Nyit, ON/OFF logikával).
- [lockd2/backend/frontend/app.js](./lockd2/backend/frontend/app.js): A beállító felület kliensoldali logikája az új `akku`/`elem` szűréssel és UI frissítéssel.
- [lockd2/config.yaml](./lockd2/config.yaml): Addon konfiguráció.
- [lockd2/run.sh](./lockd2/run.sh): Indító script bashio-val.

## Kapcsolódó Projektek
- [lockd-go2 Backend](https://github.com/MarciPain/lockd-go2): A központi backend.
- [lockd2 Mobilapp](https://github.com/MarciPain/lockd2): A Flutter alapú kliens.
