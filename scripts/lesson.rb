#!/usr/bin/env ruby

puts "This is lesson.rb"

require 'rbconfig'

machine = RbConfig::CONFIG['host_cpu'].downcase
puts "Detected architecture: #{machine}"

RbConfig::CONFIG.each do |key, value|
  puts "#{key}: #{value}"
end
