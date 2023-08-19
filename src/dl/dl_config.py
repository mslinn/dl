import os
import yaml
import dl.util as util
from pathlib import Path
from typing import List

class DLConfig:
    def __init__(self, config_path="~/dl.config") -> None:
        self.config_path = config_path
        self.config      = self.load(config_path)
        self.local       = DLConfig.get(self, dict=self.config, name='local')
        self.remotes     = DLConfig.get(self, dict=self.config, name='remotes')
        # self.disabled = DLConfig.disabled(self.local)
        # self.mp3s     = DLConfig.mp3s(self.local)
        # self.vdest    = DLConfig.vdest(self.local)
        # self.xdest    = DLConfig.xdest(self.local)

    def disabled(cls, dict) -> bool:
        return cls.get(dict, 'disabled')

    def get(cls, dict, name) -> str | bool | dict:
        if name in dict:
            return dict[name]
        return None

    def mp3s(cls, dict) -> Path:
        return Path(cls.get(dict, 'mp3s'))

    def vdest(cls, dict) -> Path:
        return Path(cls.get(dict, 'vdest'))

    def xdest(cls, dict) -> Path:
        return Path(cls.get(dict, 'xdest'))

    def active_remotes(self) -> List[dict]:
        def not_disabled(dict):
            if 'disabled' in dict:
                return not dict['disabled']
            return True

        return map(lambda x: x[0], filter(not_disabled, self.remotes.items()))

    # @return dictionary containing contents of YAML file
    def load(self, config_path="~/dl.config") -> 'DLConfig':
        self.config_file = os.path.expanduser(config_path)
        if not os.path.isfile(self.config_file):
            print(f"Error: {self.config_file} does not exist.")
            exit(1)

        if util.is_wsl():
            os.environ['win_home'] = str(util.win_home().resolve())

        with open(self.config_file, mode="rb") as file:
            self.config = yaml.safe_load(file)

        return self.config
