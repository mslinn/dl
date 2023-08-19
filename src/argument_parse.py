from dl_config import DLConfig
from textwrap import fill, dedent
from inspect import cleandoc
from os.path import expandvars
from argparse import ArgumentParser, HelpFormatter

# See https://stackoverflow.com/a/64102901/553865
class RawFormatter(HelpFormatter):
    def _fill_text(self, text, width, indent):
        return "\n".join([fill(line, width) for line in indent(dedent(text), indent).splitlines()])

class ArgParse:
    def __init__(self, config: DLConfig) -> None:
        self.action = 'mp3'
        self.config = config
        self.format = 'mp3'

        self.parser = self.make_arg_parser()
        self.args = self.parser.parse_args()

        self.action = 'mp3'
        self.format = 'mp3'
        if self.args.keep_video or self.args.video_dest or self.args.xrated:
            self.action = 'video'
            self.format = 'mp4'

    def make_description(self):
        remotes = list(self.config.active_remotes())
        aremotes = ", ".join(list(map(lambda x: x, remotes)))
        return cleandoc(f"""
            Downloads media.
            Defaults to just downloading an MP3, even when the original is a video, unless the -x, -v or -V options are provided.

            Modify {self.config.config_file} to suit; at present,
            MP3s are downloaded to {self.config.local['mp3s']},
            videos to {self.config.local['vdest']}, and
            x-rated videos to {self.config.local['xdest']}.
            Active remotes are: {aremotes}.
        """)

    def make_arg_parser(self):
        desc = self.make_description()
        desc_expanded = expandvars(desc)
        parser = ArgumentParser(prog='dl',
                        description=desc_expanded,
                        epilog=f"Last modified 2023-08-18.",
                        formatter_class=RawFormatter)
        parser.add_argument('url')
        parser.add_argument('-d', '--debug', action='store_true', help="Enable debug mode")
        parser.add_argument('-v', '--keep-video', action='store_true', help=expandvars(f"Download video to {self.config.local['vdest']} and remotes"))
        parser.add_argument('-x', '--xrated', action='store_true', help=expandvars(f"Download video to {self.config.local['xdest']} and remotes"))
        parser.add_argument('-V', '--video_dest', help=f"download video to the specified directory")
        return parser
