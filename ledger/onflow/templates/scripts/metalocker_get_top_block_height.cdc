{{ define "metalocker_get_top_block_height" }}
import MetaLocker from {{.MetaLocker}}

access(all) fun main() : UInt64 {
    return MetaLocker.getTopBlockHeight()
}
{{ end }}
