import os
import sys
import glob
import yaml

PERSIST_KEYS = {
    "persist_long_name": "mf_long_name",
    "persist_postal_code": "mf_postal_code",
    "persist_address": "mf_address",
    "persist_tel": "mf_tel",
    "persist_fax": "mf_fax",
    "persist_email": "mf_email",
    "persist_website": "mf_website",
}

def main(root: str) -> int:
    pattern = os.path.join(root, "**", "@company.yaml")
    files = glob.glob(pattern, recursive=True)
    if not files:
        print("no @company.yaml found")
        return 1

    for company_path in files:
        dir_path = os.path.dirname(company_path)
        manifest_path = os.path.join(dir_path, "@manifest.yaml")

        with open(company_path, "r", encoding="utf-8") as f:
            data = yaml.safe_load(f) or {}

        mapped = {}
        for src_key, dst_key in PERSIST_KEYS.items():
            mapped[dst_key] = data.get(src_key, "")

        with open(manifest_path, "w", encoding="utf-8") as f:
            yaml.safe_dump(mapped, f, allow_unicode=True, sort_keys=False)

        print(f"migrated: {manifest_path}")

    return 0

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("usage: python migrate_manifest.py \"C:/path/to/1 会社\"")
        sys.exit(1)
    sys.exit(main(sys.argv[1]))
