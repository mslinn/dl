import os
from os.path import expandvars
from pathlib import Path
from typing import Union
from yaml import safe_load
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

        self.active_remotes: dict = self.find_active_remotes()

    @classmethod
    def disabled(cls, dictionary: dict) -> bool:
        """Indicates if a remote is defined but disabled
        Args:
            dictionary (dict): dictionary defining remote
        Returns:
            bool: true if remote is disabled
        """
        x: StrDictNone = cls.get(dictionary, 'disabled')
        if isinstance(x, bool):
            return x
        return False

    @classmethod
    def get(cls, dictionary: dict, name: str) -> str | dict | None:
        """
        Args:
            dictionary (dict): dictionary defining all remotes
            name (str): name of dictionary to search for
        Returns:
            bool: dictionary for named remote
        """
        if name in dictionary:
            return dictionary[name]
        return None

    @classmethod
    def mp3s(cls, dictionary: dict) -> Path | None:
        """Expands environment variables in mp3 Path
        Args:
            dictionary (dict): dictionary defining remote
        Returns:
            Path if remote defines a Path for mp3s, otherwise returns None
        """
        x: StrDictNone = cls.get(dictionary, 'mp3s')
        if isinstance(x, str):
            return Path(expandvars(x))
        return None

    @classmethod
    def vdest(cls, dictionary: dict) -> Path | None:
        """Expands environment variables in video Path
        Args:
            dictionary (dict): dictionary defining remote
        Returns:
            Path if remote defines a Path for videos, otherwise returns None
        """
        x: StrDictNone = cls.get(dictionary, 'vdest')
        if isinstance(x, str):
            return Path(expandvars(x))
        return None

    @classmethod
    def xdest(cls, dictionary: dict) -> Path | None:
        """Expands environment variables in x-rated video Path
        Args:
            dictionary (dict): dictionary defining remote
        Returns:
            Path if remote defines a Path for x-rated videos, otherwise returns None
        """
        x: StrDictNone = cls.get(dictionary, 'xdest')
        if isinstance(x, str):
            return Path(expandvars(x))
        return None

    def find_active_remotes(self) -> dict:
        """
        Returns:
            dict: dictionary of dictionaries defining all enabled remotes
        """
        result: dict = {}
        if isinstance(self.remotes, dict):
            for key in self.remotes:
                remote = self.remotes[key]
                if 'disabled' in remote and remote['disabled']: continue
                result[key] = remote
        return result

    @classmethod
    def expand_entry(cls, hash: dict, name: str) -> None:
        """Updates hash with modified entry.
        Works becase Python uses call by reference.
        Args:
            hash (dict): Contains entry to be expanded
            name (str): key for entry to be expanded
        """
        x: StrDictNone = DLConfig.get(hash, name)
        if isinstance(x, str):
            hash[name] = expandvars(x)

    def load(self, config_path="~/dl.config") -> 'DLConfig':
        """Parses configuration file
        Args:
            config_path (str): points to configuration file
        Returns:
            DLConfig: dictionary containing contents of YAML file
        """
        self.config_file = os.path.expanduser(config_path)
        if not os.path.isfile(self.config_file):
            print(f"Error: {self.config_file} does not exist.")
            exit(1)

        if is_wsl():
            os.environ['win_home'] = str(win_home())

        with open(self.config_file, mode="rb") as file:
            self.config = safe_load(file)

        local = self.config['local']
        if local:
            if 'mp3s'  in local: DLConfig.expand_entry(local, 'mp3s')
            if 'vdest' in local: DLConfig.expand_entry(local, 'vdest')
            if 'xdest' in local: DLConfig.expand_entry(local, 'xdest')

        remotes = self.config['remotes']
        if remotes:
            for key in remotes:
                remote = remotes[key]
                if 'mp3s'  in remote: DLConfig.expand_entry(remote, 'mp3s')
                if 'vdest' in remote: DLConfig.expand_entry(remote, 'vdest')
                if 'xdest' in remote: DLConfig.expand_entry(remote, 'xdest')

        return self.config
