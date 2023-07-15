import json
import re
import sys
import yt_dlp

format = 'mp3'
if len(sys.argv)==1:
    URL = 'https://www.youtube.com/watch?v=BaW_jenozKc'
else:
    URL = sys.argv[1]

with yt_dlp.YoutubeDL({}) as ydl:
    info = ydl.extract_info(URL, download=False)

name = re.sub(r'[^A-Za-z0-9 ]+', '', info['title']).strip().replace(' ', '_')
name = re.sub('__+', '_', name)
ydl_opts = {
    'format': f'{format}/bestaudio/best',
    # See help(yt_dlp.postprocessor) for a list of available Postprocessors and their arguments
    'postprocessors': [{  # Extract audio using ffmpeg
        'key': 'FFmpegExtractAudio',
        'preferredcodec': format,
    }],
    'outtmpl': name
}

with yt_dlp.YoutubeDL(ydl_opts) as ydl:
    # print(json.dumps(ydl.sanitize_info(info)))
    error_code = ydl.download(URL)
