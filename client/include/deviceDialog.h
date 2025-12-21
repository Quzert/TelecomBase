#pragma once

#include <QDialog>
#include <QList>

#include "models.h"

class ApiClient;
class QComboBox;
class QLineEdit;
class QTextEdit;

class DeviceDialog final : public QDialog {
    Q_OBJECT

public:
    explicit DeviceDialog(ApiClient* apiClient, QWidget* parent = nullptr);

    bool loadReferenceData(QString& outError);
    void setInitialDevice(const DeviceDetails& device);

    qint64 selectedModelId() const;
    QVariant selectedLocationId() const;

    QString serialNumber() const;
    QString inventoryNumber() const;
    QString status() const;
    QString installedAt() const;
    QString description() const;

private:
    void buildUi();
    void applyInitialSelection();

    ApiClient* apiClient_;

    QList<ModelItem> models_;
    QList<LocationItem> locations_;

    DeviceDetails initialDevice_;
    bool hasInitialDevice_;

    QComboBox* modelCombo_;
    QComboBox* locationCombo_;
    QLineEdit* serialEdit_;
    QLineEdit* inventoryEdit_;
    QLineEdit* statusEdit_;
    QLineEdit* installedAtEdit_;
    QTextEdit* descriptionEdit_;
};
