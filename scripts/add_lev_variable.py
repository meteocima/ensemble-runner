import xarray as xr 
import os

ds = xr.open_dataset(os.environ['FILE_PATH'] + '.remapd')
ds = ds.rename({'num_press_levels_stag': 'lev'})
lev_var = xr.DataArray(ds['lev'], attrs={'axis':'Z'})
ds = ds.assign({'lev': lev_var})
ds.to_netcdf(os.environ['FILE_PATH'] + '.levfixd')
