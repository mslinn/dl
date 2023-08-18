import os
import yaml
import util

class DLConfig:
    def __init__(self) -> None:
        self.config   = self.load()
        self.local    = DLConfig.get(self, dict=self.config, name='local')
        self.remotes  = DLConfig.get(self, dict=self.config, name='remotes')
        # self.disabled = DLConfig.disabled(self.local)
        # self.mp3s     = DLConfig.mp3s(self.local)
        # self.vdest    = DLConfig.vdest(self.local)
        # self.xdest    = DLConfig.xdest(self.local)

    def disabled(cls, dict):
        return cls.get(dict, 'disabled')

    def get(cls, dict, name):
        if name in dict:
            return dict[name]
        return None

    def mp3s(cls, dict):
        return cls.get(dict, 'mp3s')

    def vdest(cls, dict):
        return cls.get(dict, 'vdest')

    def xdest(cls, dict):
        return cls.get(dict, 'xdest')

    def active_remotes(self):
        def not_disabled(dict):
            if 'disabled' in dict:
                return not dict['disabled']
            return True

        return map(lambda x: x[0], filter(not_disabled, self.remotes.items()))

    # @return dictionary containing contents of YAML file
    def load(self):
        self.config_file = os.path.expanduser("~/dl.config")
        if not os.path.isfile(self.config_file):
            os.exit(f"Error: {self.config_file} does not exist.")

        if util.is_wsl():
            os.environ['win_home'] = util.win_home()

        with open(self.config_file, mode="rb") as file:
            self.config = yaml.safe_load(file)

        return self.config
