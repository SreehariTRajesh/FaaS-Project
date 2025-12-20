import pandas as pd
import os
from scipy.stats import mannwhitneyu
directories = [
    "imagepr",
    "linpack",
    "lrserv",
    "rnnserv",
    "vidpr",
    "webserv",
    "wordcnt"
]
import csv 

summary_file = open('summary.csv', 'w+')

writer = csv.writer(summary_file)

writer.writerow(['benchmark', 'freq0', 'freq1', 'p-value'])

for dir in directories:
    benchmark = dir
    file_paths = []
    freq = []
    for file in os.listdir(dir):
        f = file[:-4][len(benchmark):]
        print(f)
        freq.append(f)
        file_path = os.path.join(dir, file)
        file_paths.append(file_path)
    
    for sfile, sfreq in zip(file_paths, freq):
        for dfile, dfreq in zip(file_paths, freq):
            sdf = pd.read_csv(sfile)
            ddf = pd.read_csv(dfile)
            u, p = mannwhitneyu(sdf.pid, ddf.pid, alternative='two-sided')
            writer.writerow([benchmark, sfreq, dfreq, p])