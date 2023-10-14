#!/bin/env python3
# set CAPTURE_DIRECTORY for watched folder
# set VMTOUCH_MAX_SIZE for max size
# set SLEEP_SECONDS for sleeping between cycles

from pathlib import Path
from os import environ, stat_result
import logging
import time
import subprocess

logging.basicConfig(format='[%(levelname)s] %(message)s', encoding='utf-8', level=logging.INFO)

CAPTURE_DIRECTORY = Path(environ.get("CAPTURE_DIRECTORY","/data"))
MAX_SIZE = environ.get("MAX_SIZE","1G")
SLEEP_SECONDS = float(environ.get("SLEEP_SECONDS",60))

cache = {}

def add(path: str):
  global cache
  proc = cache.get(path, None)
  if proc is not None and proc.poll() is None:
    return
  logging.info(f"Keeping {path} in memory.")
  proc = subprocess.Popen(["/usr/local/bin/vmtouch", "-l", path])
  cache[path] = proc

def remove(path: str):
  global cache
  proc = cache.get(path, None)
  if proc is None:
    return
  logging.info(f"Releasing {path} from memory.")
  proc.kill()
  del cache[path]

def remove_missing():
  global cache
  toremove = []
  for file in cache.keys():
    path = Path(file)
    if not path.exists():
      toremove.append(path.as_posix())
  for file in toremove:
    remove(file)

class FileWithInfo:
  path: Path
  info: stat_result

  def __init__(self, path: Path) -> None:
    self.path = path
    self.info = path.stat()
  
  def __repr__(self) -> str:
    return f"<FileWithInfo path={str(self.path)} size={self.info.st_size} mtime={self.info.st_mtime}>"

def main():
  global CAPTURE_DIRECTORY
  global MAX_SIZE
  global SLEEP_SECONDS
  try:
    sizes_dict = {'B': 1, 'K': 1024, 'M': 1024*1024, 'G': 1024*1024*1024}
    for k, v in sizes_dict.items():
      if MAX_SIZE[-1] == k:
          MAX_SIZE = int(MAX_SIZE[:-1]) * v
          break
  except:
    pass
  logging.info(f"Using max size bytes {MAX_SIZE}")

  while True:
    try:
      logging.info("Refreshing directory")
      files = []
      for file in CAPTURE_DIRECTORY.iterdir():
        try:
          if file.is_file():
            files.append( FileWithInfo(file) )
        except:
          pass
      files = sorted(files, key=lambda x: x.info.st_mtime, reverse=True)
      sum_size = 0
      for file in files:
        sum_size += file.info.st_size
        if sum_size < MAX_SIZE:
          add(file.path.as_posix())
        else:
          remove(file.path.as_posix())
      remove_missing()
    except:
      logging.exception("Error during runtime")
    logging.info(f"Sleeping {SLEEP_SECONDS} seconds")
    time.sleep(SLEEP_SECONDS)

main()