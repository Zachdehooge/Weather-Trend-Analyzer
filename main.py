from trends.dewpointplotter import dewpointplotter
from trends.dewtrendplotter import dewtrendplotter
from trends.precippointplotter import precippointplotter
from trends.preciptrendplotter import preciptrendplotter
from trends.temppointplotter import temppointplotter
from trends.temptrendplotter import temptrendplotter
from rich.console import Console

console = Console()

def main():

    console.print("1. Create Temperature Plot", style="bold red")
    console.print("2. Create Temperature Trend", style="bold red")
    console.print("3. Create Precipitation Plot", style="bold blue")
    console.print("4. Create Precipitation Trend", style="bold blue")
    console.print("5. Create Dew Point Plot", style="bold green")
    console.print("6. Create Dew Point Trend", style="bold green")
    console.print("7. Exit", style="bold yellow")

    choose = input("Enter a choice: ")

    if choose == "1":
        console.print("_____ Temperature Plot _______________________________", style="bold red")
        temppointplotter()
        main()
    if choose == "2":
        console.print("_____ Temperature Trend _______________________________", style="bold red")
        temptrendplotter()
        main()
    if choose == "3":
        console.print("_____ Precipitation Plot _______________________________", style="bold blue")
        precippointplotter()
        main()
    if choose == "4":
        console.print("_____ Precipitation Trend _______________________________", style="bold blue")
        preciptrendplotter()
        main()
    if choose == "5":
        console.print("_____ Dew Point Plot _______________________________", style="bold green")
        dewpointplotter()
        main()
    if choose == "6":
        console.print("_____ Dew Point Trend _______________________________", style="bold green")
        dewtrendplotter()
        main()
    if choose == "7":
        exit()

if __name__ == "__main__":
    main()