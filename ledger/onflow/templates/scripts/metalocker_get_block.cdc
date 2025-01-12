{{ define "metalocker_get_block" }}
import MetaLocker from {{.MetaLocker}}

access(all) fun main(number: UInt64) : MetaLocker.MetaBlock? {
    return MetaLocker.getBlock(number: number)
}
{{ end }}
