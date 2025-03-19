import matplotlib.pyplot as plt
import openmeteo_requests
import pandas as pd
import requests_cache
from retry_requests import retry

# Setup the Open-Meteo API client with cache and retry on error
cache_session = requests_cache.CachedSession('.cache', expire_after = -1)
retry_session = retry(cache_session, retries = 5, backoff_factor = 0.2)
openmeteo = openmeteo_requests.Client(session = retry_session)

# Make sure all required weather variables are listed here
# The order of variables in hourly or daily is important to assign them correctly below
url = "https://archive-api.open-meteo.com/v1/archive"
params = {
    "latitude": 35.0456,
    "longitude": -85.3097,
    "start_date": "2024-01-01",
    "end_date": "2024-06-01",
    "hourly": "temperature_2m",
    "temperature_unit": "fahrenheit",
    "wind_speed_unit": "mph"
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
hourly_temperature_2m = hourly.Variables(0).ValuesAsNumpy()

hourly_data = {"date": pd.date_range(
    start=pd.to_datetime(hourly.Time(), unit="s", utc=True),
    end=pd.to_datetime(hourly.TimeEnd(), unit="s", utc=True),
    freq=pd.Timedelta(seconds=hourly.Interval()),
    inclusive="left"
), "temperature_2m": hourly_temperature_2m}

hourly_dataframe = pd.DataFrame(data = hourly_data)
print(hourly_dataframe)

# Plot the temperature data
plt.figure(figsize=(12, 7))
plt.plot(hourly_dataframe['date'], hourly_dataframe['temperature_2m'], color='tab:red')
plt.title(f'Hourly Temperature Data', fontsize=16)
plt.xlabel('Date', fontsize=12)
plt.ylabel('Temperature (°F)', fontsize=12)
plt.grid(True, alpha=0.3)

# Format the x-axis to show dates more clearly
# Since this is just 2 days of data, show hours instead of months
# plt.gca().xaxis.set_major_formatter(mdates.DateFormatter('%b %d %H:%M'))
# plt.gca().xaxis.set_major_locator(mdates.HourLocator(interval=4))  # Show every 4 hours
# plt.xticks(rotation=35)  # Rotate date labels for better readability

# Add a horizontal line for freezing point
plt.axhline(y=32, color='blue', linestyle='--', alpha=0.7, label='Freezing Point (32°F)')

# Add date display at the bottom of the graph
plt.figtext(0.165, 0.001, f"Data period: {hourly_dataframe['date'].min().strftime('%Y-%m-%d %H:%M')} to {hourly_dataframe['date'].max().strftime('%Y-%m-%d %H:%M')}",
           ha='center', fontsize=10)

plt.tight_layout()
plt.subplots_adjust(bottom=0.15)  # Make room for the date text at the bottom
plt.legend()
plt.show()