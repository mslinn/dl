from dl_config import DLConfig
import inspect
import os
import textwrap
from argparse import ArgumentParser, HelpFormatter

class RawFormatter(HelpFormatter):
    def _fill_text(self, text, width, indent):
        return "\n".join([textwrap.fill(line, width) for line in textwrap.indent(textwrap.dedent(text), indent).splitlines()])

class ArgParse:
    def __init__(self, config: DLConfig) -> None:
        self.action = 'mp3'
        self.config = config
        self.format = 'mp3'

        parser = self.make_arg_parser()
        args = parser.parse_args()

        action = 'mp3'
        format = 'mp3'
        if args.keep_video or args.video_dest or args.xrated:
            action = 'video'
            format = 'mp4'

    def make_description(self):
        remotes = list(self.config.active_remotes())
        aremotes = ", ".join(list(map(lambda x: x, remotes)))
        return inspect.cleandoc(f"""
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
        desc_expanded = os.path.expandvars(desc)
        parser = ArgumentParser(prog='dl',
                        description=desc_expanded,
                        epilog=f"Last modified 2023-07-16.",
                        formatter_class=RawFormatter)
        parser.add_argument('url')
        parser.add_argument('-d', '--debug', action='store_true', help="Enable debug mode")
        parser.add_argument('-v', '--keep-video', action='store_true', help=os.path.expandvars(f"Download video to {self.config.local['vdest']} and remotes"))
        parser.add_argument('-x', '--xrated', action='store_true', help=os.path.expandvars(f"Download video to {self.config.local['xdest']} and remotes"))
        parser.add_argument('-V', '--video_dest', help=f"download video to the specified directory")
        return parser
