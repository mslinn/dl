import os
import yaml
import util

def not_disabled(x):
    name = x[0]
    props = x[1]
    return not 'disabled' in props or props['disabled']==False

def active_remotes():
    return map(lambda x: x[0], filter(not_disabled, config['remotes'].items()))

def read_config():
    global config, config_file
    config_file = os.path.expanduser("~/dl.config")
    if not os.path.isfile(config_file):
        os.exit(f"Error: {config_file} does not exist.")

    if util.is_wsl():
        os.environ['win_home'] = util.win_home()

    with open(config_file, mode="rb") as file:
        config = yaml.safe_load(file)
