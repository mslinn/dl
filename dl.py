import json
import re
import yt_dlp

format = 'mp3'
URL = 'https://www.youtube.com/watch?v=BaW_jenozKc'

with yt_dlp.YoutubeDL({}) as ydl:
    info = ydl.extract_info(URL, download=False)

name = re.sub(r'[^A-Za-z0-9 ]+', '', info['title']).strip().replace(' ', '_')
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
