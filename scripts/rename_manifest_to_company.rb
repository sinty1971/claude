#!/usr/bin/env ruby
# frozen_string_literal: true

require 'fileutils'
require 'pathname'

# メイン処理: @manifest.yaml を @company.yaml にリネーム
def rename_manifests(root)
  # Windowsのバックスラッシュをスラッシュに変換
  normalized_root = root.gsub('\\', '/')
  
  # @manifest.yaml ファイルを再帰的に検索
  pattern = File.join(normalized_root, '**', '@manifest.yaml')
  puts "検索パターン: #{pattern}"
  
  manifest_files = Dir.glob(pattern)

  if manifest_files.empty?
    puts '@manifest.yaml ファイルが見つかりませんでした'
    return 1
  end

  puts "#{manifest_files.length} 件の @manifest.yaml ファイルが見つかりました"
  puts ''

  renamed_count = 0
  error_count = 0

  manifest_files.each do |manifest_path|
    begin
      process_rename(manifest_path)
      renamed_count += 1
    rescue StandardError => e
      puts "エラー: #{manifest_path} - #{e.message}"
      error_count += 1
    end
  end

  puts ''
  puts "完了: #{renamed_count} 件リネーム, #{error_count} 件エラー"

  error_count.zero? ? 0 : 1
end

# 個別ファイルのリネーム処理
def process_rename(manifest_path)
  dir_path = File.dirname(manifest_path)
  company_path = File.join(dir_path, '@company.yaml')

  # すでに @company.yaml が存在する場合の処理
  if File.exist?(company_path)
    puts "スキップ: #{manifest_path} (@company.yaml が既に存在)"
    return
  end

  # ファイルをリネーム
  FileUtils.mv(manifest_path, company_path)
  
  puts "リネーム: #{manifest_path} → #{company_path}"
end

# エントリーポイント
if __FILE__ == $PROGRAM_NAME
  if ARGV.length < 1
    puts '使用方法: ruby rename_manifest_to_company.rb "C:/path/to/1 会社"'
    puts ''
    puts '例:'
    puts '  ruby scripts/rename_manifest_to_company.rb "C:/Users/user/penguin/1 会社"'
    puts '  ruby scripts/rename_manifest_to_company.rb ~/penguin/1\ 会社'
    exit 1
  end

  root_path = ARGV[0]
  
  unless Dir.exist?(root_path)
    puts "エラー: ディレクトリが存在しません: #{root_path}"
    exit 1
  end

  puts "処理開始: #{root_path}"
  puts ''
  
  exit rename_manifests(root_path)
end
