# migrate_manifest.nu
# 使い方: nu migrate_manifest.nu "c:/SyncFolder/SynologyDrive/豊田築炉/1 会社"

def main [root: string] {
    let company_files = (
        glob ($root | path join '**' '@company.yaml')
    )

    for file in $company_files {
        let dir = ($file | path dirname)
        let manifest = ($dir | path join "@manifest.yaml")

        let data = (open $file | from yaml)

        let mapped = (
            $data
            | upsert mf_long_name $data.persist_long_name
            | upsert mf_postal_code $data.persist_postal_code
            | upsert mf_address $data.persist_address
            | upsert mf_tel $data.persist_tel
            | upsert mf_fax $data.persist_fax
            | upsert mf_email $data.persist_email
            | upsert mf_website $data.persist_website
            | reject persist_long_name persist_postal_code persist_address persist_tel persist_fax persist_email persist_website
        )

        $mapped | to yaml | save -f $manifest
        print $"migrated: ($manifest)"
    }
}
