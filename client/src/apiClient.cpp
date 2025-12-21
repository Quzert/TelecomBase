#include "apiClient.h"

#include <QEventLoop>
#include <QJsonDocument>
#include <QJsonArray>
#include <QJsonObject>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QNetworkRequest>
#include <QUrlQuery>
#include <QTimer>
#include <QUrl>

ApiClient::ApiClient(QObject* parent)
    : QObject(parent)
    , baseUrl_("http://localhost:8080") {
}

void ApiClient::setBaseUrl(const QString& baseUrl) {
    baseUrl_ = baseUrl;
}

QString ApiClient::baseUrl() const {
    return baseUrl_;
}

QString ApiClient::token() const {
    return token_;
}

void ApiClient::setToken(const QString& token) {
    token_ = token;
}

ApiClient::AuthResult ApiClient::registerUser(const QString& username, const QString& password) {
    return postAuth("/auth/register", username, password);
}

ApiClient::AuthResult ApiClient::login(const QString& username, const QString& password) {
    return postAuth("/auth/login", username, password);
}

bool ApiClient::listVendors(QList<VendorItem>& outVendors, QString& outError) {
    QJsonArray arr;
    if (!getJsonArray("/vendors", arr, outError)) {
        return false;
    }

    outVendors.clear();
    for (const auto& v : arr) {
        if (!v.isObject()) {
            continue;
        }
        const QJsonObject o = v.toObject();
        VendorItem item;
        item.id = static_cast<qint64>(o.value("id").toDouble());
        item.name = o.value("name").toString();
        item.country = o.value("country").toString();
        outVendors.push_back(item);
    }
    return true;
}

bool ApiClient::createVendor(const QString& name, const QString& country, QString& outError) {
    QJsonObject body;
    body["name"] = name;
    body["country"] = country;

    QJsonObject obj;
    return postJsonObject("/vendors", body, obj, outError);
}

bool ApiClient::updateVendor(qint64 id, const QString& name, const QString& country, QString& outError) {
    QJsonObject body;
    body["name"] = name;
    body["country"] = country;

    QJsonObject obj;
    return putJsonObject(QString("/vendors/%1").arg(id), body, obj, outError);
}

bool ApiClient::deleteVendor(qint64 id, QString& outError) {
    QJsonObject obj;
    return deleteRequest(QString("/vendors/%1").arg(id), obj, outError);
}

ApiClient::AuthResult ApiClient::postAuth(const QString& path, const QString& username, const QString& password) {
    AuthResult result;

    QNetworkAccessManager manager;
    QNetworkRequest request(QUrl(baseUrl_ + path));
    request.setHeader(QNetworkRequest::ContentTypeHeader, "application/json");

    QJsonObject body;
    body["username"] = username;
    body["password"] = password;

    QNetworkReply* reply = manager.post(request, QJsonDocument(body).toJson(QJsonDocument::Compact));

    QEventLoop loop;
    QTimer timeout;
    timeout.setSingleShot(true);

    QObject::connect(&timeout, &QTimer::timeout, &loop, [&]() {
        result.ok = false;
        result.error = "timeout";
        reply->abort();
        loop.quit();
    });
    QObject::connect(reply, &QNetworkReply::finished, &loop, &QEventLoop::quit);

    timeout.start(7000);
    loop.exec();

    result.httpStatus = reply->attribute(QNetworkRequest::HttpStatusCodeAttribute).toInt();

    const QByteArray responseBytes = reply->readAll();
    const QJsonDocument responseDoc = QJsonDocument::fromJson(responseBytes);

    if (reply->error() != QNetworkReply::NoError) {
        if (responseDoc.isObject()) {
            const QJsonObject obj = responseDoc.object();
            result.error = obj.value("error").toString(reply->errorString());
        } else if (result.error.isEmpty()) {
            result.error = reply->errorString();
        }
        reply->deleteLater();
        return result;
    }

    if (!responseDoc.isObject()) {
        result.ok = false;
        result.error = "invalid_response";
        reply->deleteLater();
        return result;
    }

    const QJsonObject obj = responseDoc.object();

    if (obj.contains("token")) {
        result.ok = true;
        result.token = obj.value("token").toString();
        result.username = obj.value("username").toString();
        result.role = obj.value("role").toString();
    } else {
        result.ok = false;
        result.error = obj.value("error").toString("unknown_error");
    }

    reply->deleteLater();
    return result;
}

bool ApiClient::listModels(QList<ModelItem>& outModels, QString& outError) {
    QJsonArray arr;
    if (!getJsonArray("/models", arr, outError)) {
        return false;
    }

    outModels.clear();
    for (const auto& v : arr) {
        if (!v.isObject()) {
            continue;
        }
        const QJsonObject o = v.toObject();
        ModelItem item;
        item.id = static_cast<qint64>(o.value("id").toDouble());
        item.vendorId = static_cast<qint64>(o.value("vendorId").toDouble());
        item.vendorName = o.value("vendorName").toString();
        item.name = o.value("name").toString();
        item.deviceType = o.value("deviceType").toString();
        outModels.push_back(item);
    }

    return true;
}

bool ApiClient::createModel(qint64 vendorId, const QString& name, const QString& deviceType, QString& outError) {
    QJsonObject body;
    body["vendorId"] = static_cast<double>(vendorId);
    body["name"] = name;
    body["deviceType"] = deviceType;

    QJsonObject obj;
    return postJsonObject("/models", body, obj, outError);
}

bool ApiClient::updateModel(qint64 id, qint64 vendorId, const QString& name, const QString& deviceType, QString& outError) {
    QJsonObject body;
    body["vendorId"] = static_cast<double>(vendorId);
    body["name"] = name;
    body["deviceType"] = deviceType;

    QJsonObject obj;
    return putJsonObject(QString("/models/%1").arg(id), body, obj, outError);
}

bool ApiClient::deleteModel(qint64 id, QString& outError) {
    QJsonObject obj;
    return deleteRequest(QString("/models/%1").arg(id), obj, outError);
}

bool ApiClient::listLocations(QList<LocationItem>& outLocations, QString& outError) {
    QJsonArray arr;
    if (!getJsonArray("/locations", arr, outError)) {
        return false;
    }

    outLocations.clear();
    for (const auto& v : arr) {
        if (!v.isObject()) {
            continue;
        }
        const QJsonObject o = v.toObject();
        LocationItem item;
        item.id = static_cast<qint64>(o.value("id").toDouble());
        item.name = o.value("name").toString();
        item.note = o.value("note").toString();
        outLocations.push_back(item);
    }

    return true;
}

bool ApiClient::createLocation(const QString& name, const QString& note, QString& outError) {
    QJsonObject body;
    body["name"] = name;
    body["note"] = note;

    QJsonObject obj;
    return postJsonObject("/locations", body, obj, outError);
}

bool ApiClient::updateLocation(qint64 id, const QString& name, const QString& note, QString& outError) {
    QJsonObject body;
    body["name"] = name;
    body["note"] = note;

    QJsonObject obj;
    return putJsonObject(QString("/locations/%1").arg(id), body, obj, outError);
}

bool ApiClient::deleteLocation(qint64 id, QString& outError) {
    QJsonObject obj;
    return deleteRequest(QString("/locations/%1").arg(id), obj, outError);
}

bool ApiClient::listDevices(const QString& query, QList<DeviceItem>& outDevices, QString& outError) {
    QUrl url(baseUrl_ + "/devices");
    QUrlQuery q;
    q.addQueryItem("q", query);
    url.setQuery(q);

    QJsonArray arr;
    if (!getJsonArray(url.path() + "?" + url.query(), arr, outError)) {
        return false;
    }

    outDevices.clear();
    for (const auto& v : arr) {
        if (!v.isObject()) {
            continue;
        }
        const QJsonObject o = v.toObject();
        DeviceItem item;
        item.id = static_cast<qint64>(o.value("id").toDouble());
        item.vendorName = o.value("vendorName").toString();
        item.modelName = o.value("modelName").toString();
        item.locationName = o.value("locationName").toString();
        item.serialNumber = o.value("serialNumber").toString();
        item.inventoryNumber = o.value("inventoryNumber").toString();
        item.status = o.value("status").toString();
        item.installedAt = o.value("installedAt").toString();
        outDevices.push_back(item);
    }

    return true;
}

bool ApiClient::getDevice(qint64 id, DeviceDetails& outDevice, QString& outError) {
    QJsonObject obj;
    if (!getJsonObject(QString("/devices/%1").arg(id), obj, outError)) {
        return false;
    }

    outDevice = DeviceDetails{};
    outDevice.id = static_cast<qint64>(obj.value("id").toDouble());
    outDevice.modelId = static_cast<qint64>(obj.value("modelId").toDouble());

    if (obj.contains("locationId") && !obj.value("locationId").isNull()) {
        outDevice.locationId = obj.value("locationId").toVariant();
    } else {
        outDevice.locationId = QVariant();
    }

    outDevice.serialNumber = obj.value("serialNumber").toString();
    outDevice.inventoryNumber = obj.value("inventoryNumber").toString();
    outDevice.status = obj.value("status").toString();
    outDevice.installedAt = obj.value("installedAt").toString();
    outDevice.description = obj.value("description").toString();
    return true;
}

bool ApiClient::createDevice(qint64 modelId, const QVariant& locationId, const QString& serialNumber,
                             const QString& inventoryNumber, const QString& status, const QString& installedAt,
                             const QString& description, QString& outError) {
    QJsonObject body;
    body["modelId"] = static_cast<double>(modelId);
    if (locationId.isValid() && !locationId.isNull()) {
        body["locationId"] = locationId.toLongLong();
    }
    body["serialNumber"] = serialNumber;
    body["inventoryNumber"] = inventoryNumber;
    body["status"] = status;
    body["installedAt"] = installedAt;
    body["description"] = description;

    QJsonObject obj;
    return postJsonObject("/devices", body, obj, outError);
}

bool ApiClient::updateDevice(qint64 id, qint64 modelId, const QVariant& locationId, const QString& serialNumber,
                             const QString& inventoryNumber, const QString& status, const QString& installedAt,
                             const QString& description, QString& outError) {
    QJsonObject body;
    body["modelId"] = static_cast<double>(modelId);
    if (locationId.isValid() && !locationId.isNull()) {
        body["locationId"] = locationId.toLongLong();
    }
    body["serialNumber"] = serialNumber;
    body["inventoryNumber"] = inventoryNumber;
    body["status"] = status;
    body["installedAt"] = installedAt;
    body["description"] = description;

    QJsonObject obj;
    return putJsonObject(QString("/devices/%1").arg(id), body, obj, outError);
}

bool ApiClient::deleteDevice(qint64 id, QString& outError) {
    QJsonObject obj;
    return deleteRequest(QString("/devices/%1").arg(id), obj, outError);
}

bool ApiClient::listUsers(QList<UserItem>& outUsers, QString& outError) {
    QJsonArray arr;
    if (!getJsonArray("/users", arr, outError)) {
        return false;
    }

    outUsers.clear();
    for (const auto& v : arr) {
        if (!v.isObject()) {
            continue;
        }
        const QJsonObject o = v.toObject();
        UserItem item;
        item.id = static_cast<qint64>(o.value("id").toDouble());
        item.username = o.value("username").toString();
        item.role = o.value("role").toString();
        item.approved = o.value("approved").toBool(false);
        item.createdAt = o.value("createdAt").toString();
        outUsers.push_back(item);
    }

    return true;
}

bool ApiClient::setUserApproved(qint64 id, bool approved, QString& outError) {
    QJsonObject body;
    body["approved"] = approved;

    QJsonObject obj;
    return putJsonObject(QString("/users/%1/approval").arg(id), body, obj, outError);
}

bool ApiClient::deleteUser(qint64 id, QString& outError) {
    QJsonObject obj;
    return deleteRequest(QString("/users/%1").arg(id), obj, outError);
}

bool ApiClient::getJsonArray(const QString& path, QJsonArray& outArray, QString& outError) {
    QNetworkAccessManager manager;
    QNetworkRequest request(QUrl(baseUrl_ + path));
    request.setHeader(QNetworkRequest::ContentTypeHeader, "application/json");
    if (!token_.isEmpty()) {
        request.setRawHeader("Authorization", ("Bearer " + token_).toUtf8());
    }

    QNetworkReply* reply = manager.get(request);

    QEventLoop loop;
    QTimer timeout;
    timeout.setSingleShot(true);

    QObject::connect(&timeout, &QTimer::timeout, &loop, [&]() {
        outError = "timeout";
        reply->abort();
        loop.quit();
    });
    QObject::connect(reply, &QNetworkReply::finished, &loop, &QEventLoop::quit);

    timeout.start(7000);
    loop.exec();

    const QByteArray responseBytes = reply->readAll();
    const QJsonDocument doc = QJsonDocument::fromJson(responseBytes);

    if (reply->error() != QNetworkReply::NoError) {
        if (outError.isEmpty()) {
            outError = reply->errorString();
        }
        reply->deleteLater();
        return false;
    }
    if (!doc.isArray()) {
        outError = "invalid_response";
        reply->deleteLater();
        return false;
    }

    outArray = doc.array();
    reply->deleteLater();
    return true;
}

bool ApiClient::getJsonObject(const QString& path, QJsonObject& outObject, QString& outError) {
    QNetworkAccessManager manager;
    QNetworkRequest request(QUrl(baseUrl_ + path));
    request.setHeader(QNetworkRequest::ContentTypeHeader, "application/json");
    if (!token_.isEmpty()) {
        request.setRawHeader("Authorization", ("Bearer " + token_).toUtf8());
    }

    QNetworkReply* reply = manager.get(request);

    QEventLoop loop;
    QTimer timeout;
    timeout.setSingleShot(true);

    QObject::connect(&timeout, &QTimer::timeout, &loop, [&]() {
        outError = "timeout";
        reply->abort();
        loop.quit();
    });
    QObject::connect(reply, &QNetworkReply::finished, &loop, &QEventLoop::quit);

    timeout.start(7000);
    loop.exec();

    const QByteArray responseBytes = reply->readAll();
    const QJsonDocument doc = QJsonDocument::fromJson(responseBytes);

    if (reply->error() != QNetworkReply::NoError) {
        if (doc.isObject()) {
            outError = doc.object().value("error").toString(reply->errorString());
        } else {
            outError = reply->errorString();
        }
        reply->deleteLater();
        return false;
    }
    if (!doc.isObject()) {
        outError = "invalid_response";
        reply->deleteLater();
        return false;
    }

    outObject = doc.object();
    reply->deleteLater();
    return true;
}

bool ApiClient::postJsonObject(const QString& path, const QJsonObject& body, QJsonObject& outObject, QString& outError) {
    QNetworkAccessManager manager;
    QNetworkRequest request(QUrl(baseUrl_ + path));
    request.setHeader(QNetworkRequest::ContentTypeHeader, "application/json");
    if (!token_.isEmpty()) {
        request.setRawHeader("Authorization", ("Bearer " + token_).toUtf8());
    }

    QNetworkReply* reply = manager.post(request, QJsonDocument(body).toJson(QJsonDocument::Compact));

    QEventLoop loop;
    QTimer timeout;
    timeout.setSingleShot(true);

    QObject::connect(&timeout, &QTimer::timeout, &loop, [&]() {
        outError = "timeout";
        reply->abort();
        loop.quit();
    });
    QObject::connect(reply, &QNetworkReply::finished, &loop, &QEventLoop::quit);

    timeout.start(7000);
    loop.exec();

    const QByteArray responseBytes = reply->readAll();
    const QJsonDocument doc = QJsonDocument::fromJson(responseBytes);

    if (reply->error() != QNetworkReply::NoError) {
        if (doc.isObject()) {
            outError = doc.object().value("error").toString(reply->errorString());
        } else {
            outError = reply->errorString();
        }
        reply->deleteLater();
        return false;
    }
    if (!doc.isObject()) {
        outError = "invalid_response";
        reply->deleteLater();
        return false;
    }

    outObject = doc.object();
    reply->deleteLater();
    return true;
}

bool ApiClient::putJsonObject(const QString& path, const QJsonObject& body, QJsonObject& outObject, QString& outError) {
    QNetworkAccessManager manager;
    QNetworkRequest request(QUrl(baseUrl_ + path));
    request.setHeader(QNetworkRequest::ContentTypeHeader, "application/json");
    if (!token_.isEmpty()) {
        request.setRawHeader("Authorization", ("Bearer " + token_).toUtf8());
    }

    QNetworkReply* reply = manager.put(request, QJsonDocument(body).toJson(QJsonDocument::Compact));

    QEventLoop loop;
    QTimer timeout;
    timeout.setSingleShot(true);

    QObject::connect(&timeout, &QTimer::timeout, &loop, [&]() {
        outError = "timeout";
        reply->abort();
        loop.quit();
    });
    QObject::connect(reply, &QNetworkReply::finished, &loop, &QEventLoop::quit);

    timeout.start(7000);
    loop.exec();

    const QByteArray responseBytes = reply->readAll();
    const QJsonDocument doc = QJsonDocument::fromJson(responseBytes);

    if (reply->error() != QNetworkReply::NoError) {
        if (doc.isObject()) {
            outError = doc.object().value("error").toString(reply->errorString());
        } else {
            outError = reply->errorString();
        }
        reply->deleteLater();
        return false;
    }
    if (!doc.isObject()) {
        outError = "invalid_response";
        reply->deleteLater();
        return false;
    }

    outObject = doc.object();
    reply->deleteLater();
    return true;
}

bool ApiClient::deleteRequest(const QString& path, QJsonObject& outObject, QString& outError) {
    QNetworkAccessManager manager;
    QNetworkRequest request(QUrl(baseUrl_ + path));
    request.setHeader(QNetworkRequest::ContentTypeHeader, "application/json");
    if (!token_.isEmpty()) {
        request.setRawHeader("Authorization", ("Bearer " + token_).toUtf8());
    }

    QNetworkReply* reply = manager.deleteResource(request);

    QEventLoop loop;
    QTimer timeout;
    timeout.setSingleShot(true);

    QObject::connect(&timeout, &QTimer::timeout, &loop, [&]() {
        outError = "timeout";
        reply->abort();
        loop.quit();
    });
    QObject::connect(reply, &QNetworkReply::finished, &loop, &QEventLoop::quit);

    timeout.start(7000);
    loop.exec();

    const QByteArray responseBytes = reply->readAll();
    const QJsonDocument doc = QJsonDocument::fromJson(responseBytes);

    if (reply->error() != QNetworkReply::NoError) {
        if (doc.isObject()) {
            outError = doc.object().value("error").toString(reply->errorString());
        } else {
            outError = reply->errorString();
        }
        reply->deleteLater();
        return false;
    }

    if (doc.isObject()) {
        outObject = doc.object();
    }

    reply->deleteLater();
    return true;
}
