# Use Cases — schliz/deckel

Hier sind die Anwendungsfälle, die diese Software implementiert, niedergeschrieben. Die Kennzeichnungen lauten wiefolgt:
- [ ] offen: Keine Testfälle
- [~] teilweise getestet
- [x] vollständig getestet (insb. e2e)

## Authentifizierung & Sitzung

- [ ] Erstanmeldung: Ein neuer Benutzer meldet sich über oauth2-proxy an. Das System legt automatisch ein Benutzerkonto mit Name und E-Mail an. Der Benutzer sieht die Getränkekarte.
- [ ] Wiederkehrende Anmeldung: Ein bestehender Benutzer meldet sich an. Seine Profildaten (Name, Admin-Status) werden aktualisiert. Er sieht die Getränkekarte mit seinem aktuellen Guthaben.
- [ ] Deaktivierter Benutzer: Ein deaktivierter Benutzer meldet sich an. Das System blockiert den Zugriff mit einer Fehlermeldung (403).
- [ ] Abmeldung: Ein Benutzer klickt auf "Abmelden" in der Navigation. Er wird über oauth2-proxy abgemeldet.

## Getränkekarte & Bestellung

- [x] Karte ansehen: Ein Benutzer öffnet die Getränkekarte. Er sieht alle Kategorien mit deren Getränken und den für seinen Status (Barteamer/Helfer) gültigen Preisen.
- [x] Getränk bestellen: Ein Benutzer klickt auf ein Getränk, wählt im Modal die Menge (1 bis Maximalwert aus Einstellungen) und bestätigt die Bestellung. Das Guthaben verringert sich um den Gesamtbetrag und eine Erfolgsmeldung erscheint.
- [ ] Bestellung bei niedrigem Guthaben: Ein Benutzer bestellt ein Getränk und unterschreitet dabei das Warnlimit. Die Bestellung wird durchgeführt, aber es erscheint eine Warnung ("Ausgabelimit erreicht — bitte bald einzahlen!").
- [ ] Bestellung bei Ausgabelimit: Ein Benutzer hat das harte Ausgabelimit erreicht. Die Getränke auf der Karte sind ausgegraut und nicht klickbar. Bestellungen werden serverseitig abgelehnt.
- [ ] Bestellung bei deaktiviertem Limit: Ein Benutzer, dessen Ausgabelimit vom Admin aufgehoben wurde, kann trotz negativem Guthaben weiterhin bestellen.

## Eigene Buchung

- [x] Eigene Buchung erstellen: Ein Benutzer klickt auf "Eigene Buchung", gibt eine Beschreibung und einen Betrag (innerhalb der konfigurierten Grenzen) ein und bestätigt. Eine Transaktion wird erstellt und das Guthaben entsprechend angepasst.
- Eigene Buchung bei Ausgabelimit: Ein Benutzer mit erreichtem Ausgabelimit kann keine eigene Buchung erstellen — der Button ist deaktiviert.

## Transaktionshistorie

- [x] Transaktionen ansehen: Ein Benutzer öffnet die Transaktionsseite. Er sieht seine Transaktionen (Datum, Beschreibung, Menge, Betrag) paginiert und nach Datum sortiert (neueste zuerst).
- [x] Transaktion stornieren: Ein Benutzer klickt auf den Stornieren-Button einer Transaktion innerhalb des Stornierungszeitraums, bestätigt im Modal. Die Transaktion wird als storniert markiert (durchgestrichen) und eine Gegenbuchung erstellt. Das Guthaben wird wiederhergestellt.
- [ ] Stornierungsfenster abgelaufen: Ein Benutzer sieht bei älteren Transaktionen keinen Stornieren-Button, da das Zeitfenster (konfigurierbar in Minuten) abgelaufen ist.
- [ ] Stornierung einer Stornierung: Stornierungstransaktionen selbst können nicht erneut storniert werden.

## Profil

- [x] Profil ansehen: Ein Benutzer öffnet die Profilseite. Er sieht seinen Namen, seine E-Mail, seinen Status (Barteamer/Helfer), das Mitgliedsdatum und sein aktuelles Guthaben.
- [ ] Daten exportieren: Ein Benutzer klickt auf "Daten exportieren". Er erhält eine ZIP-Datei mit seinen Profildaten (JSON) und seiner Transaktionshistorie (CSV).

## Header & Navigation

- [x] Guthabenanzeige: Nach dem Laden jeder Seite sieht der Benutzer in der Kopfzeile sein Guthaben, das Gesamtguthaben aller Benutzer und seinen Rang. Admins sehen zusätzlich die Summe aller negativen Guthaben.
- Warnhinweis bei niedrigem Guthaben: Ein Benutzer mit niedrigem Guthaben sieht einen Warnhinweis-Banner unterhalb der Navigation.
- [x] Navigation: Ein Benutzer sieht die Navigationslinks Karte, Transaktionen und Profil. Ein Admin sieht zusätzlich das Admin-Untermenü (Karte, Nutzer, Transaktionen, Statistiken, Einstellungen).
- [x] Themenumschaltung: Ein Benutzer klickt auf das Sonnen-/Mond-Icon. Das Farbschema wechselt zwischen Hell und Dunkel und wird im Browser gespeichert.

## Admin: Kartenverwaltung

- [x] Kategorie erstellen: Ein Admin gibt auf der Kartenverwaltungsseite einen Kategorienamen ein und bestätigt. Die Kategorie erscheint in der Liste.
- [x] Kategorie umbenennen: Ein Admin klickt auf "Bearbeiten" bei einer Kategorie, ändert den Namen im Modal und bestätigt. Der Name wird aktualisiert.
- [x] Kategorie löschen: Ein Admin klickt auf "Löschen" bei einer leeren Kategorie und bestätigt. Die Kategorie wird entfernt. Kategorien mit Getränken können nicht gelöscht werden.
- Kategorie sortieren: Ein Admin klickt auf die Hoch-/Runter-Pfeile einer Kategorie. Die Reihenfolge der Kategorien ändert sich entsprechend.
- [x] Getränk hinzufügen: Ein Admin gibt in einer Kategorie Name, Barteamer-Preis und Helfer-Preis ein und bestätigt. Das Getränk erscheint in der Kategorie.
- [x] Getränk bearbeiten: Ein Admin klickt auf "Bearbeiten" bei einem Getränk, ändert Name und/oder Preise im Modal und bestätigt. Die Änderungen werden übernommen.
- [x] Getränk entfernen: Ein Admin klickt auf "Löschen" bei einem Getränk und bestätigt. Das Getränk wird als gelöscht markiert (Soft-Delete) und verschwindet von der Karte, bleibt aber in der Transaktionshistorie erhalten.
- [ ] Getränk sortieren: Ein Admin klickt auf die Hoch-/Runter-Pfeile eines Getränks. Die Reihenfolge innerhalb der Kategorie ändert sich.

## Admin: CSV-Stapelbearbeitung

- [ ] CSV exportieren: Ein Admin öffnet die Stapelbearbeitungsseite, wählt optional eine Kategorie und klickt "Exportieren". Er erhält eine CSV-Datei (Semikolon-getrennt) mit allen Getränken (ID, Kategorie, Name, Barteamer-Preis, Helfer-Preis).
- [ ] CSV hochladen und Vorschau: Ein Admin lädt eine bearbeitete CSV-Datei hoch. Das System zeigt eine Vorschau der Änderungen (Diff) mit Hervorhebung der geänderten Felder und eventuellen Warnungen an.
- [ ] Änderungen anwenden: Ein Admin prüft die Vorschau und klickt "Anwenden". Alle Änderungen werden transaktional übernommen. Ein Ergebnis mit Erfolgsmeldung wird angezeigt.
- [ ] CSV mit Fehlern: Ein Admin lädt eine CSV mit ungültigen Daten hoch (falsches Format, fehlende Spalten, ungültige Preise). Das System zeigt Validierungsfehler an, und keine Änderungen werden übernommen.

## Admin: Benutzerverwaltung

- [x] Benutzerliste ansehen: Ein Admin öffnet die Benutzerverwaltungsseite. Er sieht alle Benutzer paginiert, sortiert nach Guthaben (aufsteigend), mit Name, E-Mail, Guthaben, Status und Aktionen.
- [x] Einzahlung buchen: Ein Admin klickt auf "Einzahlung" bei einem Benutzer, gibt einen Betrag und optional eine Notiz ein und bestätigt. Das Guthaben des Benutzers erhöht sich um diesen Betrag.
- [x] Benutzer deaktivieren: Ein Admin klickt auf den Aktiv-Toggle eines Benutzers und bestätigt im Modal. Der Benutzer wird deaktiviert und kann sich nicht mehr anmelden. Ein Admin kann sich nicht selbst deaktivieren.
- [x] Benutzer aktivieren: Ein Admin klickt auf den Aktiv-Toggle eines deaktivierten Benutzers und bestätigt. Der Benutzer kann sich wieder anmelden.
- [x] Barteamer-Status umschalten: Ein Admin klickt auf den Barteamer-Toggle und bestätigt. Der Benutzer wechselt zwischen Barteamer und Helfer, was seine Getränkepreise beeinflusst.
- [ ] Ausgabelimit aufheben: Ein Admin klickt auf den Limit-Override-Toggle und bestätigt. Der Benutzer ist vom harten Ausgabelimit befreit und kann trotz negativem Guthaben bestellen.
- [ ] Ausgabelimit wiederherstellen: Ein Admin klickt erneut auf den Limit-Override-Toggle und bestätigt. Das Ausgabelimit gilt wieder für den Benutzer.

## Admin: Transaktionsverwaltung

- [ ] Alle Transaktionen ansehen: Ein Admin öffnet die Transaktionsverwaltung. Er sieht alle Transaktionen aller Benutzer paginiert mit Datum, Benutzer, Beschreibung, Betrag, Typ und Status.
- [ ] Transaktion stornieren (Admin): Ein Admin klickt auf "Stornieren" bei einer Transaktion innerhalb des Zeitfensters und bestätigt. Die Transaktion wird storniert und eine Gegenbuchung erstellt.

## Admin: Statistiken

- [x] Statistiken ansehen: Ein Admin öffnet die Statistikseite. Er sieht für den aktuellen Monat: Gesamtumsatz, Einzahlungen, Transaktionsanzahl, Top-10-Getränke nach Anzahl, Top-10-Getränke nach Umsatz und Umsatz nach Kategorie.
- [x] Zeitraum filtern: Ein Admin wählt einen vordefinierten Zeitraum (Heute, Letzte Woche, Letzter Monat) oder gibt einen benutzerdefinierten Datumsbereich ein. Die Statistiken werden für den gewählten Zeitraum aktualisiert.

## Admin: Einstellungen

- [x] Einstellungen ansehen: Ein Admin öffnet die Einstellungsseite. Er sieht alle konfigurierbaren Werte in Gruppen: Limits, Buchungs- und Bestelloptionen, SMTP-Konfiguration und E-Mail-Vorlage.
- [x] Einstellungen speichern: Ein Admin ändert einen oder mehrere Werte und klickt "Speichern". Die Einstellungen werden aktualisiert und wirken sich sofort auf das System aus.
- [ ] Erinnerungs-E-Mails senden: Ein Admin klickt auf "Erinnerungen senden" und bestätigt. Das System sendet an alle aktiven Benutzer eine E-Mail mit ihrem aktuellen Guthaben gemäß der konfigurierten Vorlage.

## Kiosk-Modus

- [ ] Kiosk-Anmeldung: Ein Kiosk-Benutzer meldet sich an. Er wird automatisch zur Kiosk-Ansicht weitergeleitet.
- [ ] Getränk am Kiosk bestellen: Am Kiosk wird ein Getränk ausgewählt, dann ein Benutzer. Eine Bestätigungsseite zeigt Getränk, Benutzer, Preis und Guthaben. Nach Bestätigung wird die Buchung erstellt.
- [ ] Kiosk-Bestellung bei niedrigem Guthaben: Eine Bestellung wird für einen Benutzer mit niedrigem Guthaben am Kiosk durchgeführt. Die Buchung wird erstellt, aber eine Warnung erscheint.
- [ ] Kiosk-Bestellung bei Ausgabelimit: Am Kiosk wird ein Benutzer mit erreichtem Ausgabelimit ausgewählt. Die Bestätigung ist nicht möglich, ein Hinweis erscheint.
- [ ] Kiosk-Stornierung: Am Kiosk wird eine selbst erstellte Buchung innerhalb des Zeitfensters storniert. Die Transaktion wird rückgängig gemacht.
- [ ] Kiosk-Stornierung fremder Buchungen: Am Kiosk kann nur eine vom Kiosk erstellte Buchung storniert werden, nicht die Buchungen der Benutzer selbst.
