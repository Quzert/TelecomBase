#pragma once

#include <QString>

struct VendorItem {
    qint64 id = 0;
    QString name;
    QString country;

    QString displayName() const {
        return name;
    }
};

struct ModelItem {
    qint64 id = 0;
    qint64 vendorId = 0;
    QString vendorName;
    QString name;
    QString deviceType;

    QString displayName() const {
        if (vendorName.isEmpty()) {
            return name;
        }
        return vendorName + " — " + name;
    }
};

struct LocationItem {
    qint64 id = 0;
    QString name;
    QString note;

    QString displayName() const {
        return name;
    }
};

struct DeviceItem {
    qint64 id = 0;
    QString vendorName;
    QString modelName;
    QString locationName;
    QString serialNumber;
    QString inventoryNumber;
    QString status;
    QString installedAt;
};

struct DeviceDetails {
    qint64 id = 0;
    qint64 modelId = 0;
    QVariant locationId;
    QString serialNumber;
    QString inventoryNumber;
    QString status;
    QString installedAt;
    QString description;
};

struct UserItem {
    qint64 id = 0;
    QString username;
    QString role;
    bool approved = false;
    QString createdAt;
};
