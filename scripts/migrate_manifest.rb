#!/usr/bin/env ruby
# frozen_string_literal: true

require 'yaml'
require 'pathname'

# キーマッピング定義
PERSIST_KEYS = {
  'persist_long_name' => 'mf_long_name',
  'persist_postal_code' => 'mf_postal_code',
  'persist_address' => 'mf_address',
  'persist_tel' => 'mf_tel',
  'persist_fax' => 'mf_fax',
  'persist_email' => 'mf_email',
  'persist_website' => 'mf_website'
}.freeze

# メイン処理
def migrate_manifests(root)
  # Windowsのバックスラッシュをスラッシュに変換
  normalized_root = root.gsub('\\', '/')
  # @company.yaml ファイルを再帰的に検索
  pattern = File.join(normalized_root, '**', '@company.yaml')
  puts "searching for: #{pattern}"
  company_files = Dir.glob(pattern)

  if company_files.empty?
    puts 'no @company.yaml found'
    return 1
  end

  company_files.each do |company_path|
    process_company_file(company_path)
  end

  0
end

# 個別ファイル処理
def process_company_file(company_path)
  dir_path = File.dirname(company_path)
  manifest_path = File.join(dir_path, '@manifest.yaml')

  # @company.yaml を読み込み
  data = YAML.load_file(company_path) || {}

  # キーをマッピング
  mapped = {}
  PERSIST_KEYS.each do |src_key, dst_key|
    mapped[dst_key] = data[src_key] || ''
  end

  # @manifest.yaml に書き込み
  File.write(manifest_path, YAML.dump(mapped))

  puts "migrated: #{manifest_path}"
end

# エントリーポイント
if __FILE__ == $PROGRAM_NAME
  if ARGV.length < 1
    puts 'usage: ruby migrate_manifest.rb "C:/path/to/1 会社"'
    exit 1
  end

  exit migrate_manifests(ARGV[0])
end
