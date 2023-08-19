import os
from os.path import expandvars
from yaml import safe_load
from pathlib import Path
from typing import List, Union
from dl.util import is_wsl, win_home

StrDictNone = Union[str, dict, None]

class DLConfig:
    def __init__(self, config_path="~/dl.config") -> None:
        self.config_path: str = config_path
        self.config: DLConfig = self.load(config_path)

        x: StrDictNone = DLConfig.get(dict=self.config, name='local')
        self.local: dict | None = x if isinstance(x, dict) else None

        x = DLConfig.get(dict=self.config, name='remotes')
        self.remotes: dict | None = x if isinstance(x, dict) else None

    @classmethod
    def disabled(cls, dict) -> bool:
        x: StrDictNone = cls.get(dict, 'disabled')
        if isinstance(x, bool):
            return x
        return False

    @classmethod
    def get(cls, dict, name) -> str | dict | None:
        if name in dict:
            return dict[name]
        return None

    @classmethod
    def mp3s(cls, dict) -> Path | None:
        x: StrDictNone = cls.get(dict, 'mp3s')
        if isinstance(x, str):
            return Path(expandvars(x))
        return None

    @classmethod
    def vdest(cls, dict) -> Path | None:
        x: StrDictNone = cls.get(dict, 'vdest')
        if isinstance(x, str):
            return Path(expandvars(x))
        return None

    @classmethod
    def xdest(cls, dict) -> Path | None:
        x: StrDictNone = cls.get(dict, 'xdest')
        if isinstance(x, str):
            return Path(expandvars(x))
        return None

    def active_remotes(self) -> List[dict]:
        def not_disabled(dict):
            if 'disabled' in dict:
                return not dict['disabled']
            return True

        return list(map(lambda x: x[0], filter(not_disabled, self.remotes.items()))) \
            if isinstance(self.remotes, dict) else list()

    # @return dictionary containing contents of YAML file
    def load(self, config_path="~/dl.config") -> 'DLConfig':
        self.config_file = os.path.expanduser(config_path)
        if not os.path.isfile(self.config_file):
            print(f"Error: {self.config_file} does not exist.")
            exit(1)

        if is_wsl():
            os.environ['win_home'] = str(win_home())

        with open(self.config_file, mode="rb") as file:
            self.config = safe_load(file)

        return self.config
