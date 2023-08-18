import config
import inspect
import os
import textwrap
from argparse import ArgumentParser, HelpFormatter

class RawFormatter(HelpFormatter):
    def _fill_text(self, text, width, indent):
        return "\n".join([textwrap.fill(line, width) for line in textwrap.indent(textwrap.dedent(text), indent).splitlines()])

def parse_args():
    global action, dl_options, format
    aremotes = ", ".join(list(map(lambda x: x, config.active_remotes())))
    description = inspect.cleandoc(f"""
        Downloads media.
        Defaults to just downloading an MP3, even when the original is a video, unless the -x, -v or -V options are provided.

        Modify {config.config_file} to suit; at present,
        MP3s are downloaded to {config.config['local']['mp3s']},
        videos to {config.config['local']['vdest']}, and
        x-rated videos to {config.config['local']['xdest']}.
        Active remotes are: {aremotes}.
    """)
    parser = ArgumentParser(prog='dl',
                            description=os.path.expandvars(description),
                            epilog=f"Last modified 2023-07-16.",
                            formatter_class=RawFormatter)
    parser.add_argument('url')
    parser.add_argument('-d', '--debug', action='store_true', help="Enable debug mode")
    parser.add_argument('-v', '--keep-video', action='store_true', help=os.path.expandvars(f"Download video to {config.config['local']['vdest']} and remotes"))
    parser.add_argument('-x', '--xrated', action='store_true', help=os.path.expandvars(f"Download video to {config.config['local']['xdest']} and remotes"))
    parser.add_argument('-V', '--video_dest', help=f"download video to the specified directory")
    args = parser.parse_args()

    action = 'mp3'
    format = 'mp3'
    if args.keep_video or args.video_dest or args.xrated:
        action = 'video'
        format = 'mp4'
    return args
