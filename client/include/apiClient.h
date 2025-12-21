#pragma once

#include <QObject>
#include <QString>
#include <QVariant>

#include "models.h"

class ApiClient final : public QObject {
    Q_OBJECT

public:
    struct AuthResult {
        bool ok = false;
        QString token;
        QString username;
        QString role;
        QString error;
        int httpStatus = 0;
    };

    explicit ApiClient(QObject* parent = nullptr);

    void setBaseUrl(const QString& baseUrl);
    QString baseUrl() const;

    QString token() const;
    void setToken(const QString& token);

    AuthResult registerUser(const QString& username, const QString& password);
    AuthResult login(const QString& username, const QString& password);

    bool listVendors(QList<VendorItem>& outVendors, QString& outError);
    bool createVendor(const QString& name, const QString& country, QString& outError);
    bool updateVendor(qint64 id, const QString& name, const QString& country, QString& outError);
    bool deleteVendor(qint64 id, QString& outError);

    bool listModels(QList<ModelItem>& outModels, QString& outError);
    bool createModel(qint64 vendorId, const QString& name, const QString& deviceType, QString& outError);
    bool updateModel(qint64 id, qint64 vendorId, const QString& name, const QString& deviceType, QString& outError);
    bool deleteModel(qint64 id, QString& outError);

    bool listLocations(QList<LocationItem>& outLocations, QString& outError);
    bool createLocation(const QString& name, const QString& note, QString& outError);
    bool updateLocation(qint64 id, const QString& name, const QString& note, QString& outError);
    bool deleteLocation(qint64 id, QString& outError);
    bool listDevices(const QString& query, QList<DeviceItem>& outDevices, QString& outError);

    bool getDevice(qint64 id, DeviceDetails& outDevice, QString& outError);

    bool createDevice(qint64 modelId, const QVariant& locationId, const QString& serialNumber,
                      const QString& inventoryNumber, const QString& status, const QString& installedAt,
                      const QString& description, QString& outError);

    bool updateDevice(qint64 id, qint64 modelId, const QVariant& locationId, const QString& serialNumber,
                      const QString& inventoryNumber, const QString& status, const QString& installedAt,
                      const QString& description, QString& outError);
    bool deleteDevice(qint64 id, QString& outError);

    bool listUsers(QList<UserItem>& outUsers, QString& outError);
    bool setUserApproved(qint64 id, bool approved, QString& outError);
    bool deleteUser(qint64 id, QString& outError);

private:
    AuthResult postAuth(const QString& path, const QString& username, const QString& password);

    bool getJsonArray(const QString& path, QJsonArray& outArray, QString& outError);
    bool getJsonObject(const QString& path, QJsonObject& outObject, QString& outError);
    bool postJsonObject(const QString& path, const QJsonObject& body, QJsonObject& outObject, QString& outError);
    bool putJsonObject(const QString& path, const QJsonObject& body, QJsonObject& outObject, QString& outError);
    bool deleteRequest(const QString& path, QJsonObject& outObject, QString& outError);

    QString baseUrl_;
    QString token_;
};
