{{ define "metalocker_import_block" }}
import MetaLocker from {{.MetaLocker}}

transaction(number: UInt64, records: [MetaLocker.Record]) {
    execute {
        MetaLocker.importBlock(number: number, records: records)
    }
}
{{ end }}
