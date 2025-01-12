{{ define "metalocker_get_asset_head" }}
import MetaLocker from {{.MetaLocker}}

access(all) fun main(headID: String) : MetaLocker.Record? {
    return MetaLocker.getAssetHead(headID: headID)
}
{{ end }}
