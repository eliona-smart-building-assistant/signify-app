# Signify User Guide

### Introduction

> The Signify app provides integration and synchronization between Eliona and [Signify services](https://www.signify.com/).

## Overview

This guide provides instructions on configuring, installing, and using the Signify app to manage resources and synchronize data between Eliona and Signify services.

## Installation

Install the Signify app via the Eliona App Store.

## Configuration

The Signify app requires configuration through Eliona’s settings interface. Below are the general steps and details needed to configure the app effectively.

### Registering the app in Signify Service

Create credentials in Signify Service to connect the Signify services from Eliona. All required credentials are listed below in the [configuration section](#configure-the-signify-app).  

Go to the [Interact Lighting website](https://www.developer.interact-lighting.com/) and create an Eliona client to obtain the app's key and secret needed to access the API.
Additionally, you must identify the service you want to connect to, as this information is provided by Interact Lighting.

### Configure the Signify app 

Configurations can be created in Eliona under `Apps > Signify > Settings` which opens the app's [Generic Frontend](https://doc.eliona.io/collection/v/eliona-english/manuals/settings/apps). Here you can use the appropriate endpoint with the POST method. Each configuration requires the following data:

| Attribute         | Description                                                                              |
|-------------------|------------------------------------------------------------------------------------------|
| `baseURL`         | URL of the Signify [API services](https://www.developer.interact-lighting.com/api-docs). |
| `app_key`         | The app key to identify the ELiona app.                                                  |
| `app_secret`      | The app secret to authenticate the Elioa app.                                            |
| `service`         | The service name to identify the Interact Lighting service.                              |
| `service_id`      | The service id to identify the Interact Lighting service.                                |
| `service_secret`  | The service secret to authenticate the Interact Lighting service.                        |
| `assetFilter`     | Filtering asset during [Continuous Asset Creation](#continuous-asset-creation).          |
| `enable`          | Flag to enable or disable this configuration.                                            |
| `refreshInterval` | Interval in seconds for data synchronization.                                            |
| `requestTimeout`  | API query timeout in seconds.                                                            |
| `projectIDs`      | List of Eliona project IDs for data collection.                                          |

Example configuration JSON:

```json
{
  "baseUrl": "https://api.interact-lighting.com/",
  "service": "officeCloud",
  "serviceId": "secret@signify.com",
  "serviceSecret": "secret",
  "appKey": "foo",
  "appSecret": "secret",
  "enable": true,
  "refreshInterval": 120,
  "requestTimeout": 120,
  "active": true,
  "projectIDs": [
    "10"
  ]
}
```

## Continuous Asset Creation

Once configured, the app starts Continuous Asset Creation (CAC). Discovered resources are automatically created as assets in Eliona, and users are notified via Eliona’s notification system.

The created asset structure reflects the grouping of spaces in the Interact Lighting Environment. The app supports Signify spaces for humidity, occupancy, people count and temperature. 

## Additional Features

### Dashboard templates

The app offers predefined dashboards that clearly displays the most important information.
There is a dashboard for each type Signify space to group all the spaces.
You can create such a dashboard under `Dashboards > Copy Dashboard > From App > Signify ...`:

- `Signify People Count`
- `Signify Occupancy`
- `Signify Temperature`
- `Signify Humidity`
