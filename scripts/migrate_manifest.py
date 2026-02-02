#!/usr/bin/env python3
"""migrate_manifest.py

Ruby スクリプト migrate_manifest.rb の Python 置換。

使い方:
  python scripts/migrate_manifest.py "C:/path/to/1 会社"

依存:
  pip install pyyaml
"""
import sys
from pathlib import Path
import yaml

# 古い変換テーブルは不要になったため削除。
# 処理方針:
# - 入力のキーが `persist_...` の場合は出力を `pr_...` にする
# - 入力のキーが `mf_...` の場合は出力を `pr_...` にする
def migrate_manifests(root: str, yaml_filename: str = '@company.yaml') -> int:
    print("yaml_filename:", yaml_filename)

    # Windows のバックスラッシュをスラッシュに変換
    normalized_root = root.replace('\\', '/')
    pattern = f"{normalized_root}/**/{yaml_filename}"
    print(f"searching for: {pattern}")

    base = Path(normalized_root)
    yaml_file_paths = list(base.rglob(yaml_filename)) if base.exists() else []

    if not yaml_file_paths:
        print(f'no {yaml_filename} found')
        return 1

    for yaml_file_path in yaml_file_paths:
        process_yaml_file(yaml_file_path)

    return 0


def process_yaml_file(yaml_path: Path) -> None:
    try:
        with yaml_path.open('r', encoding='utf-8') as f:
            data = yaml.safe_load(f) or {}
    except Exception as e:
        print(f"failed to read {yaml_path}: {e}")
        data = {}

    mapped = {}
    for src_key, value in data.items():
        if not src_key or value is None:
            continue

        # persist_... -> pr_...
        if src_key.startswith('persist_'):
            suffix = src_key[len('persist_'):]
            dst_key = f'pr_{suffix}'
            mapped[dst_key] = value
            continue

        # mf_... -> pr_...
        if src_key.startswith('mf_'):
            suffix = src_key[len('mf_'):]
            dst_key = f'pr_{suffix}'
            mapped[dst_key] = value
            continue

    # 既定のキーが存在しない場合に空文字列で埋めたい場合はここで追加可能

    try:
        with yaml_path.open('w', encoding='utf-8') as f:
            yaml.dump(mapped, f, allow_unicode=True, sort_keys=False)
        print(f"migrated: {yaml_path}")
    except Exception as e:
        print(f"failed to write {yaml_path}: {e}")


def main(argv):
    if len(argv) < 2:
        print('usage: python migrate_manifest.py "C:/path/to/1 会社" [yaml_filename]')
        return 1

    print("starting migration...", flush=True)
    root = argv[1]
    yaml_filename = argv[2] if len(argv) > 2 else '@company.yaml'
    return migrate_manifests(root, yaml_filename)


if __name__ == '__main__':
    try:
        sys.exit(main(sys.argv) or 0)
    except Exception:
        import traceback
        print('Unhandled exception in migrate_manifest.py:', file=sys.stderr)
        traceback.print_exc(file=sys.stderr)
        sys.exit(1)
