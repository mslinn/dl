from argparse import ArgumentParser, HelpFormatter
from inspect import cleandoc
from os.path import expandvars
from textwrap import fill, dedent, indent

# See https://stackoverflow.com/a/64102901/553865
class RawFormatter(HelpFormatter):
    def _fill_text(self, text, width, indentr):
        text = dedent(text)          # Strip the indent from the original python definition that plagues most of us.
        text = indent(text, indentr)  # Apply any requested indent.
        text = text.splitlines()              # Make a list of lines
        text = [fill(line, width) for line in text] # Wrap each line
        text = "\n".join(text)                # Join the lines again
        return text

class ArgParse:
    def __init__(self, config: 'DLConfig') -> None: # type: ignore
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

    def make_description(self) -> str:
        remotes = list(self.config.active_remotes)
        aremotes = ", ".join(list(map(lambda x: x, remotes)))
        return cleandoc(f"""
            Downloads media.
            Defaults to just downloading an MP3, even when the original is a video, unless the -x, -v or -V options are provided.

            Modify {self.config.config_file} to suit; at present,
            MP3s are downloaded to {self.config.mp3s(self.config.local)},
            videos to {self.config.vdest(self.config.local)}, and
            x-rated videos to {self.config.xdest(self.config.local)}.
            Active remotes are: {aremotes}.
        """)

    def make_arg_parser(self) -> ArgumentParser:
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
