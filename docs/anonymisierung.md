# Anonymisierung von Transaktionen

Verfahrensdokumentations-Basis für die Anonymisierung personenbezogener Daten in Transaktionen sowie für die Löschung von Benutzerkonten. Richtet sich an Admins und an Prüfer, die das Verfahren nachvollziehen müssen.

## Zweck

Deckel führt für jedes Mitglied ein internes Verzehrkonto. Die GoBD (§ 147 AO, § 257 HGB) verlangen die Aufbewahrung von Geschäftsvorfällen (Datum, Artikel, Menge, Einzelpreis, Betrag) über mehrere Jahre. Sie verlangen jedoch **nicht**, dass diese Daten dauerhaft einer Person zugeordnet bleiben — bei einem rein internen Verzehrkonto ohne Rechnungen auf Namen ist die Identität des Konsumenten nicht Teil des steuerlich relevanten Belegs (analog § 146 AO, der beim Barverkauf keinen Kundennamen fordert).

Die DSGVO verlangt umgekehrt Speicherbegrenzung (Art. 5 Abs. 1 lit. e): Personenbezogene Daten dürfen nur so lange gespeichert werden, wie der Zweck es erfordert. Der Zweck des Personenbezugs ist hier die Forderungsverwaltung (offene Salden). Ist eine Transaktion beglichen und ausreichend lange her, entfällt dieser Zweck, und der Personenbezug wird entfernt.

Anonymisierung statt Löschung erlaubt es, beide Anforderungen gleichzeitig zu erfüllen: Der Geschäftsvorfall bleibt für die Aufbewahrungsfrist vollständig erhalten, nur die Zuordnung zu einer Person entfällt.

## Was entfernt wird, was erhalten bleibt

Beim Anonymisieren einer Transaktion werden ausschließlich die beiden personenbezogenen Fremdschlüssel entfernt:

- `transactions.user_id` → `NULL`
- `transactions.created_by_user_id` → `NULL`
- `transactions.anonymized_at` wird auf den Zeitpunkt des Laufs gesetzt (Nachweis, wann anonymisiert wurde)

Alle Buchungsdaten bleiben unverändert erhalten: Betrag (`amount`), Artikel (`item_title`), Menge (`quantity`), Einzelpreis (`unit_price`), Typ (`type`), Buchungsdatum (`created_at`) und Beschreibung (`description`). Die Zeile bleibt in Statistiken (`internal/store/stats.go`) und im GoBD-Sinn vollständig nachvollziehbar — nur ohne erkennbare Person.

Ein `CHECK`-Constraint (`transactions_user_or_anonymized`) stellt sicher, dass `user_id` ausschließlich über den Anonymisierungspfad `NULL` werden kann (nie über einen anderen Codepfad).

## Saldo-Erhalt

Anonymisierung darf den Kontostand eines Mitglieds nicht verändern. Dazu trägt `users.balance_carryover` (BIGINT, Default 0) die Summe aller bei diesem Mitglied bereits anonymisierten Beträge. Beim Anonymisieren wird der Betrag der betroffenen Zeilen atomar auf `balance_carryover` aufaddiert, bevor `user_id` entfernt wird. Alle Saldo-Abfragen (`internal/store/user.go`) berücksichtigen `balance_carryover + SUM(amount)` der verbleibenden, noch zugeordneten Transaktionen — der ausgewiesene Kontostand ändert sich durch einen Lauf also nicht.

Synthetische Ausgleichsbuchungen wurden bewusst nicht gewählt: Sie würden in `GetTotalRevenue` als zusätzlicher Umsatz gezählt und als Fake-Buchungen in Historie und DSGVO-Export auftauchen.

## Stornopaar-Regel

Eine ursprüngliche Buchung und ihre Stornobuchung (verknüpft über `cancels_id`) bilden ein Paar. Ein Paar wird nur **gemeinsam** anonymisiert: Eine Zeile ist fällig, wenn sie selbst älter als der Stichtag ist **und** ihre Partnerzeile (in beide Richtungen von `cancels_id`) ebenfalls älter als der Stichtag ist. Dadurch bleibt eine kürzlich stornierte, aber ursprünglich alte Buchung so lange zusammen mit ihrem Storno zugeordnet, bis auch das Storno alt genug ist. Die maximale Verzögerung entspricht dem konfigurierten Stornierungsfenster (`cancellation_minutes`).

## Auslösung

Anonymisierung ist eine bewusste Admin-Aktion, kein automatischer Hintergrundjob:

1. Admin öffnet die Einstellungsseite und klickt auf "Anonymisierung ausführen".
2. Eine Vorschau (`Preview`) zeigt die Anzahl betroffener Transaktionen und Nutzer sowie den Stichtag, ohne Daten zu ändern.
3. Nach Bestätigung führt `Run` den Lauf in einer Transaktion aus (`internal/service/anonymize`), abgesichert durch `pg_advisory_xact_lock`, um doppelte Gutschriften bei parallelen Läufen zu verhindern.

Die Frist ist über `settings.anonymize_after_months` konfigurierbar (0–120 Monate, Default 36). Der Wert 0 deaktiviert die Anonymisierung vollständig — der Ausführen-Button wird ausgeblendet und durch einen Hinweis ersetzt.

Der Service (`internal/service/anonymize`) ist bewusst ohne HTTP-Kontext gehalten. Die Architektur erlaubt es, dieselbe Funktion später zusätzlich aus einem Hintergrundjob (z. B. periodische Goroutine in `cmd/server`) mit `executedBy = "system"` aufzurufen, ohne den Kern zu ändern. Ein solcher automatischer Lauf ist aktuell nicht implementiert.

## Protokollierung

Jeder Lauf mit mindestens einer fälligen Transaktion wird in der Tabelle `anonymization_runs` protokolliert:

| Spalte | Bedeutung |
|---|---|
| `executed_at` | Zeitpunkt der Ausführung |
| `cutoff_date` | Verwendeter Stichtag |
| `transaction_count` | Anzahl anonymisierter Transaktionen |
| `user_count` | Anzahl betroffener Nutzer |
| `executed_by` | E-Mail-Snapshot des ausführenden Admins (bewusst kein Fremdschlüssel auf `users`, damit auch spätere Systemläufe mit `'system'` protokollierbar sind) |

Läufe ohne fällige Transaktionen erzeugen keinen Protokolleintrag. Die letzten Läufe werden auf der Einstellungsseite als Tabelle angezeigt (sichtbares GoBD-Log).

## Benutzerlöschung

Ergänzend zur Anonymisierung von Transaktionen können Benutzerkonten ehemaliger Mitglieder gelöscht werden. Voraussetzungen (alle müssen erfüllt sein):

- Der Benutzer ist deaktiviert (`is_active = false`).
- Der Kontostand (inklusive `balance_carryover`) beträgt genau 0.
- Es existieren keine Transaktionen mehr, die auf den Benutzer verweisen (weder als `user_id` noch als `created_by_user_id`) — d. h. alle seine Buchungen wurden bereits anonymisiert.

Sind diese Voraussetzungen nicht erfüllt, zeigt das Lösch-Modal einen entsprechenden Hinweis statt eines Bestätigen-Buttons. Die Löschung selbst ist ein hartes `DELETE` auf die `users`-Zeile; die Guards werden beim Löschen erneut innerhalb einer Transaktion mit `FOR UPDATE` geprüft, um ein Wettrennen mit einer gerade eintreffenden Buchung auszuschließen.

**SSO-Hinweis:** Die Authentifizierungs-Middleware (`internal/auth`) provisioniert bei jedem Request automatisch ein Benutzerkonto, falls keines existiert. Meldet sich ein gelöschter Benutzer erneut über den Identity Provider an, erhält er dadurch ein frisches, leeres Konto. Ein vollständiges Offboarding erfordert zusätzlich das Entfernen bzw. Deaktivieren des Kontos beim IdP (z. B. Keycloak) — das ist außerhalb von Deckels Verantwortung.

## Auswirkung auf Historie und Export

Anonymisierte Transaktionen verlieren ihre Zuordnung zum Benutzer und verschwinden dadurch aus dessen Mitglieder-Historie und aus dem DSGVO-Datenexport (der Link zum Benutzer existiert schlicht nicht mehr). Das ist beabsichtigt. Der Kontostand des Mitglieds bleibt davon unberührt, da der Betrag zuvor über `balance_carryover` gesichert wurde.
