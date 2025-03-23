{{ define "metalocker_get_data_asset_counter" }}
import MetaLocker from {{.MetaLocker}}

access(all) fun main(id: String) : UInt64? {
    return MetaLocker.getDataAssetCounter(id: id)
}
{{ end }}
