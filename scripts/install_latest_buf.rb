#!/usr/bin/env ruby
require 'open-uri'
require 'fileutils'
require 'optparse'

def detect_arch
  machine = RbConfig::CONFIG['host_cpu'].downcase
  return 'arm64' if machine.include?('arm64') || machine.include?('aarch64')
  return 'x86_64' if machine.include?('x86_64') || machine.include?('amd64') || machine.include?('x64')

  raise "Unsupported architecture: #{machine}"
end

def build_url(version, arch)
  if version == 'latest'
    url = "https://github.com/bufbuild/buf/releases/latest/download/buf-Windows-#{arch}.exe"
    return [url, 'latest']
  end

  tag = version.start_with?('v') ? version : "v#{version}"
  url = "https://github.com/bufbuild/buf/releases/download/#{tag}/buf-Windows-#{arch}.exe"
  [url, tag]
end

def main
  options = {
    version: 'latest',
    destination: File.join(Dir.home, '.local', 'bin', 'buf.exe')
  }

  OptionParser.new do |opts|
    opts.on('--version VERSION') { |v| options[:version] = v }
    opts.on('--destination PATH') { |d| options[:destination] = d }
  end.parse!

  arch = detect_arch
  url, tag = build_url(options[:version], arch)
  dest = options[:destination]
  FileUtils.mkdir_p(File.dirname(dest))

  puts "Downloading buf (#{arch}) from #{url} ..."
  URI.open(url, 'User-Agent' => 'pathist-buf-downloader') do |source|
    File.open(dest, 'wb') { |file| file.write(source.read) }
  end

  size = File.size(dest)
  puts "Downloaded version tag: #{tag}"
  puts "Saved to #{dest} (#{size} bytes)"
end

main if __FILE__ == $PROGRAM_NAME
