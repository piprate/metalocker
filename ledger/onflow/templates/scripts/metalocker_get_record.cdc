{{ define "metalocker_get_record" }}
import MetaLocker from {{.MetaLocker}}

access(all) fun main(id: String) : MetaLocker.Record? {
    return MetaLocker.getRecord(id: id)
}
{{ end }}
