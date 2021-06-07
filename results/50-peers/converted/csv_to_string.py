import csv

with open("BlockProd.csv") as delay:
    csv_reader = csv.reader(delay, delimiter = ",")
    csv_string = ""
    for row in csv_reader:
        for index, value in enumerate(row):
            if index == len(row)-1:
                csv_string += value + "\\n"
            else:
                csv_string += value + ","
    print(csv_string)
