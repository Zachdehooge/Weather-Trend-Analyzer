import openmeteo_requests
import requests_cache
import pandas as pd
from retry_requests import retry
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
import numpy as np
from scipy import signal

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
    start = pd.to_datetime(hourly.Time(), unit = "s", utc = True),
    end = pd.to_datetime(hourly.TimeEnd(), unit = "s", utc = True),
    freq = pd.Timedelta(seconds = hourly.Interval()),
    inclusive = "left"
)}

hourly_data["temperature_2m"] = hourly_temperature_2m

hourly_dataframe = pd.DataFrame(data = hourly_data)
print(hourly_dataframe)

# Create a daily average dataframe to reduce data points
daily_data = hourly_dataframe.resample('D', on='date').mean()
daily_data = daily_data.reset_index()

# Apply a smoothing filter to get the trend
window_size = 14  # 14-day smoothing window
temp_trend = signal.savgol_filter(daily_data['temperature_2m'], window_size, 3)

# Plot the temperature trend
plt.figure(figsize=(12, 7))

# Show the trend line only (no individual points)
plt.plot(daily_data['date'], temp_trend, color='tab:red', linewidth=3, label='Temperature Trend')

plt.title(f'Temperature Trend', fontsize=16)
plt.xlabel('Date', fontsize=12)
plt.ylabel('Temperature (°F)', fontsize=12)
plt.grid(True, alpha=0.3)

# Format the x-axis to show dates more clearly
# Major ticks for months
plt.gca().xaxis.set_major_formatter(mdates.DateFormatter('%b'))
plt.gca().xaxis.set_major_locator(mdates.MonthLocator())

# Minor ticks for days (every 15 days)
plt.gca().xaxis.set_minor_locator(mdates.DayLocator(bymonthday=[1, 15]))
plt.gca().xaxis.set_minor_formatter(mdates.DateFormatter('%d'))

# Rotate dates for better readability
plt.gcf().autofmt_xdate()

# Add date display at the bottom of the graph
plt.figtext(0.5, 0.01, f"Data from: {daily_data['date'].min().strftime('%Y-%m-%d')} to {daily_data['date'].max().strftime('%Y-%m-%d')}",
            ha='center', fontsize=10)

# Add a horizontal line for freezing point
plt.axhline(y=32, color='blue', linestyle='--', alpha=0.7, label='Freezing Point (32°F)')

plt.tight_layout()
plt.subplots_adjust(bottom=0.15)
plt.legend()
plt.show()