#!/bin/env python3
# pip3 install asyncinotify requests 
# set CAPTURE_DIRECTORY for watched folder

from pathlib import Path
from asyncinotify import Inotify, Mask
from os import environ
import asyncio
import logging
import requests
import subprocess


logging.basicConfig(format='[%(levelname)s] %(message)s', encoding='utf-8', level=logging.INFO)

watch_dir = Path(environ.get("CAPTURE_DIRECTORY","/data"))

suricata_dir = Path(watch_dir, "suricata")
arkime_dir = Path(watch_dir, "arkime")

elastic_host = environ.get("ELASTIC_HOST", None)
elastic_mappings_file = environ.get("ELASTIC_MAPPINGS_FILE", "./mapping.json")

background_tasks = set()

def create_background_task(coro):
    task = asyncio.create_task(coro)
    background_tasks.add(task)
    task.add_done_callback(background_tasks.discard)

async def add_suricata_task(path: Path):
    create_background_task(arkime_delayed(path))
    suricata_path = Path(suricata_dir, path.name)
    arkime_path = Path(arkime_dir, path.name)
    if not suricata_path.exists() and not arkime_path.exists():
        logging.info(f"Adding {path.as_posix()} to Suricata.")
        suricata_path.hardlink_to(path)

async def add_arkime_task(path: Path):
    arkime_path = Path(arkime_dir, path.name)
    if not arkime_path.exists():
        await add_elastic_task(path)
        logging.info(f"Adding {path.as_posix()} to Arkime.")
        arkime_path.hardlink_to(path)
        with arkime_path.open('ab'):
            pass

async def add_elastic_task(path: Path):
    p = path.as_posix()
    pid = subprocess.Popen(f"(tshark -T ek -J 'http tcp udp ip' -x -r '{p}' | ./tshark-to-elastic '{elastic_host}/packets_template/_bulk' '{p}') &", shell=True, start_new_session=True)
    logging.info(f"Adding {p} to Elastic.")

async def arkime_delayed(path: Path):
    await asyncio.sleep(60)
    logging.info(f"Arkime timer expired for {path.as_posix()}.")
    await add_arkime_task(path)

def add_missing():
    sub = watch_dir.glob('*')
    for path in sub:
        if path.suffix == ".pcap" and path.is_file():
            suricata_path = Path(suricata_dir, path.name)
            arkime_path = Path(arkime_dir, path.name)
            if not suricata_path.exists() and not arkime_path.exists():
                create_background_task(add_suricata_task(path))

async def main():
    logging.info("Starting.")
    if elastic_host:
        for i in range(10):
            try:
                resp = requests.post(f"{elastic_host}/_cluster/health?wait_for_status=yellow", timeout=3)
                if resp.status == 200:
                    break
                logging.info("Waiting for ES to start.")
                asyncio.sleep(3000)
            except Exception as ex:
                pass
        try:
            mappings = open(elastic_mappings_file, "rb").read()
            requests.post(f"{elastic_host}/_index_template/packets_template", data=mappings, headers={"Content-Type" : "application/json"}, timeout=3)
            logging.info("Updating elastic mapping.")
        except Exception as ex:
            logging.warning(f"Error sending mapping to elastic: {ex}")
    suricata_dir.mkdir(parents=True, exist_ok=True)
    arkime_dir.mkdir(parents=True, exist_ok=True)
    with Inotify() as inotify:
        inotify.add_watch(watch_dir, Mask.CLOSE)
        inotify.add_watch(suricata_dir, Mask.DELETE)
        add_missing()
        async for event in inotify:
            if event.path:
                path = event.path
                watch = event.watch
                if path.suffix == ".pcap":
                    if watch.path == watch_dir and path.is_file():
                        create_background_task(add_suricata_task(path))
                    elif watch.path == suricata_dir:
                        watch_path = Path(watch_dir, path.name)
                        create_background_task(add_arkime_task(watch_path))
    asyncio.gather(*background_tasks);

try:
    asyncio.run(main())
except KeyboardInterrupt:
    logging.info("Shutting down.")