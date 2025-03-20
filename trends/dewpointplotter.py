import os
from datetime import datetime

import matplotlib.pyplot as plt
import openmeteo_requests
import pandas as pd
import requests
import requests_cache
from retry_requests import retry
from dotenv import load_dotenv
import numpy as np

load_dotenv()

authkey = os.getenv("APIKEY")


def dewpointplotter():
    # Setup the Open-Meteo API client with cache and retry on error
    cache_session = requests_cache.CachedSession('.cache', expire_after=-1)
    retry_session = retry(cache_session, retries=5, backoff_factor=0.2)
    openmeteo = openmeteo_requests.Client(session=retry_session)

    city = input("Enter City: ")
    state = input("Enter State: ")

    sdate = input("Enter Start Date (YYYY-MM-DD): ")
    edate = input("Enter End Date (YYYY-MM-DD): ")

    date_format = "%Y-%m-%d"

    d1 = datetime.strptime(sdate, date_format)
    d2 = datetime.strptime(edate, date_format)

    days_difference = abs((d1 - d2).days)

    if days_difference > 31:
        print("Date is out of range")
        dewpointplotter()

    base_url = "https://geocode.xyz"
    params = {
        "locate": (city + " " + state),
        "region": "US",
        "json": "1"
    }

    req_url = f"{base_url}/?{requests.utils.unquote(requests.compat.urlencode(params))}"
    try:
        resp = requests.get(req_url + f"&auth={authkey}")
        resp.raise_for_status()
    except requests.RequestException as err:
        print("Error:", err)
        exit()

    geocode_data = resp.json()

    url = "https://archive-api.open-meteo.com/v1/archive"
    params = {
        "latitude": geocode_data['latt'],
        "longitude": geocode_data['longt'],
        "start_date": sdate,
        "end_date": edate,
        "hourly": "dew_point_2m",
        "temperature_unit": "fahrenheit",
        "wind_speed_unit": "mph",
        "precipitation_unit": "inch"
    }
    responses = openmeteo.weather_api(url, params=params)

    # Process first location. Add a for-loop for multiple locations or weather models
    response = responses[0]
    print(f"Coordinates {response.Latitude()}°N {response.Longitude()}°E")
    print(f"Elevation {response.Elevation()} m asl")
    print(f"Timezone {response.Timezone()}{response.TimezoneAbbreviation()}")
    print(f"Timezone difference to GMT+0 {response.UtcOffsetSeconds()} s")

    # Process hourly data. The order of variables needs to be the same as requested.
    hourly = response.Hourly()
    hourly_dew_point_2m = hourly.Variables(0).ValuesAsNumpy()

    hourly_data = {"date": pd.date_range(
        start=pd.to_datetime(hourly.Time(), unit="s", utc=True),
        end=pd.to_datetime(hourly.TimeEnd(), unit="s", utc=True),
        freq=pd.Timedelta(seconds=hourly.Interval()),
        inclusive="left"
    ), "dew_point_2m": hourly_dew_point_2m}

    hourly_dataframe = pd.DataFrame(data=hourly_data)
    print(hourly_dataframe)

    # Plot the temperature data
    fig, ax = plt.subplots(figsize=(12, 7))

    # Plot the main line
    line, = ax.plot(hourly_dataframe['date'], hourly_dataframe['dew_point_2m'],
                    color='tab:blue', label='Dew Point', linewidth=1.5)

    # Calculate appropriate interval for marker points based on the data length
    # For example, show a point every 12 hours or once per day
    total_hours = len(hourly_dataframe)

    # Determine appropriate marker interval
    if total_hours <= 48:  # Less than 2 days of data
        marker_interval = 6  # Every 6 hours
    elif total_hours <= 168:  # Less than 1 week
        marker_interval = 12  # Every 12 hours
    else:
        marker_interval = 24  # Every 24 hours (once per day)

    # Create marker indices
    marker_indices = range(0, total_hours, marker_interval)

    # Plot markers and add annotations
    for idx in marker_indices:
        # Get the data for this marker
        date = hourly_dataframe['date'].iloc[idx]
        value = hourly_dataframe['dew_point_2m'].iloc[idx]

        # Add a marker point
        ax.plot(date, value, 'ro', ms=6, alpha=0.8, zorder=3)

        # Determine text position (alternate above/below to avoid overlap)
        if idx % 2 == 0:
            xytext = (0, 15)  # Text above point
        else:
            xytext = (0, -25)  # Text below point

        # Add text annotation with the value
        ax.annotate(f'{value:.1f}°F\n{date.strftime("%m-%d %H:%M")}',
                    xy=(date, value),
                    xytext=xytext,
                    textcoords='offset points',
                    ha='center',
                    bbox=dict(boxstyle='round,pad=0.3', fc='yellow', alpha=0.7),
                    arrowprops=dict(arrowstyle='->', connectionstyle='arc3,rad=0'))

    ax.set_title(f'Hourly Dew Point Data for {city}, {state}', fontsize=16)
    ax.set_xlabel('Date', fontsize=12)
    ax.set_ylabel('Dew Point (°F)', fontsize=12)
    ax.grid(True, alpha=0.3)

    # Add date display at the bottom of the graph
    plt.figtext(0.5, 0.01,
                f"Data period: {hourly_dataframe['date'].min().strftime('%Y-%m-%d %H:%M')} to {hourly_dataframe['date'].max().strftime('%Y-%m-%d %H:%M')}",
                ha='center', fontsize=10)

    plt.tight_layout()
    plt.subplots_adjust(bottom=0.15)  # Make room for the date text at the bottom
    plt.legend()
    plt.show()
