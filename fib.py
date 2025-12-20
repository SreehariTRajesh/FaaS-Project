import time 

def main():
    start_time = time.time()
    sum = 0
    for i in range(1, 1000000000):
        sum+=i
    end_time = time.time()
    print(end_time-start_time)


if __name__ == '__main__':
    main()

    