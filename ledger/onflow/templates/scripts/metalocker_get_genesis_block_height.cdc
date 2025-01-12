{{ define "metalocker_get_genesis_block_height" }}
import MetaLocker from {{.MetaLocker}}

access(all) fun main() : UInt64 {
    return MetaLocker.getGenesisBlockHeight()
}
{{ end }}
