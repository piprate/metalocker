{{ define "metalocker_get_top_block_number" }}
import MetaLocker from {{.MetaLocker}}

access(all) fun main() : UInt64 {
    return MetaLocker.getTopBlock()
}
{{ end }}
