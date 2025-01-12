{{ define "metalocker_submit_record" }}
import MetaLocker from {{.MetaLocker}}

transaction(rec: MetaLocker.Record) {
    execute {
        MetaLocker.submitRecord(record: rec)
    }
}
{{ end }}
