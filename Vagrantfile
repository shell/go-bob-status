# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/trusty64"

  config.vm.box_check_update = false
  config.vm.hostname = "go-box"
  config.vm.synced_folder ".", "/vagrant"
  config.vm.network "private_network", ip: "192.168.33.10"
end
