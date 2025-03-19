from trends.precippointplotter import precippointplotter
from trends.preciptrendplotter import preciptrendplotter
from trends.temppointplotter import temppointplotter
from trends.temptrendplotter import temptrendplotter

def main():

    choose = input("""
    1. Create Temperature Weather Plot
    2. Create Temperature Weather Trend
    3. Create Precipitation Weather Plot
    4. Create Precipitation Weather Trend
    5. Create Humidity Weather Plot
    6. Create Humidity Weather Trend 
    7. Create All Weather Plot
    8. Create All Weather Trend
    9. Exit
    
    Enter a choice: """)

    if choose == "1":
        print("_____ Temperature Weather Plot _______________________________")
        temppointplotter()
        main()
    if choose == "2":
        print("_____ Temperature Weather Trend _______________________________")
        temptrendplotter()
        main()
    if choose == "3":
        print("_____ Precipitation Weather Plot _______________________________")
        precippointplotter()
        main()
    if choose == "4":
        print("_____ Precipitation Weather Trend _______________________________")
        preciptrendplotter()
        main()
    if choose == "9":
        exit()

if __name__ == "__main__":
    main()