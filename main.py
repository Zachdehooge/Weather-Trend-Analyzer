# TODO: Create a choice for the user to pass location, dates, and what type of graph they want
from pointplotter import pointplotter
from trendplotter import trendplotter

def main():

    choose = input("""
    1. Create Weather Plot
    2. Create Weather Trend
    3. Exit
    
    Enter a choice: 
    """)

    if choose == "1":
        print("_____  Weather Plot _______________________________")
        pointplotter()
        main()
    if choose == "2":
        print("_____ Weather Trend _______________________________")
        trendplotter()
        main()
    if choose == "3":
        exit()

if __name__ == "__main__":
    main()