#include "deviceDialog.h"

#include "apiClient.h"

#include <QComboBox>
#include <QDialogButtonBox>
#include <QFormLayout>
#include <QLabel>
#include <QLineEdit>
#include <QTextEdit>
#include <QVBoxLayout>

DeviceDialog::DeviceDialog(ApiClient* apiClient, QWidget* parent)
    : QDialog(parent)
    , apiClient_(apiClient)
    , hasInitialDevice_(false)
    , modelCombo_(nullptr)
    , locationCombo_(nullptr)
    , serialEdit_(nullptr)
    , inventoryEdit_(nullptr)
    , statusEdit_(nullptr)
    , installedAtEdit_(nullptr)
    , descriptionEdit_(nullptr) {
    setWindowTitle("Устройство");
    buildUi();
}

void DeviceDialog::buildUi() {
    QVBoxLayout* root = new QVBoxLayout(this);

    QLabel* hint = new QLabel("Заполните поля устройства. Дата: YYYY-MM-DD (можно пусто).", this);
    hint->setWordWrap(true);
    root->addWidget(hint);

    QFormLayout* form = new QFormLayout();

    modelCombo_ = new QComboBox(this);
    form->addRow("Модель:", modelCombo_);

    locationCombo_ = new QComboBox(this);
    form->addRow("Локация:", locationCombo_);

    serialEdit_ = new QLineEdit(this);
    form->addRow("Серийный:", serialEdit_);

    inventoryEdit_ = new QLineEdit(this);
    form->addRow("Инвентарный:", inventoryEdit_);

    statusEdit_ = new QLineEdit(this);
    statusEdit_->setText("active");
    form->addRow("Статус:", statusEdit_);

    installedAtEdit_ = new QLineEdit(this);
    installedAtEdit_->setPlaceholderText("YYYY-MM-DD");
    form->addRow("Дата установки:", installedAtEdit_);

    descriptionEdit_ = new QTextEdit(this);
    descriptionEdit_->setMinimumHeight(100);
    form->addRow("Описание:", descriptionEdit_);

    root->addLayout(form);

    QDialogButtonBox* buttons = new QDialogButtonBox(QDialogButtonBox::Ok | QDialogButtonBox::Cancel, this);
    connect(buttons, &QDialogButtonBox::accepted, this, &QDialog::accept);
    connect(buttons, &QDialogButtonBox::rejected, this, &QDialog::reject);
    root->addWidget(buttons);

    setMinimumWidth(520);
}

bool DeviceDialog::loadReferenceData(QString& outError) {
    if (!apiClient_) {
        outError = "api_client_not_set";
        return false;
    }

    QList<ModelItem> models;
    if (!apiClient_->listModels(models, outError)) {
        return false;
    }

    QList<LocationItem> locations;
    if (!apiClient_->listLocations(locations, outError)) {
        return false;
    }

    models_ = models;
    locations_ = locations;

    modelCombo_->clear();
    for (const auto& m : models_) {
        modelCombo_->addItem(m.displayName(), QVariant::fromValue<qlonglong>(m.id));
    }

    locationCombo_->clear();
    locationCombo_->addItem("—", QVariant());
    for (const auto& l : locations_) {
        locationCombo_->addItem(l.displayName(), QVariant::fromValue<qlonglong>(l.id));
    }

    applyInitialSelection();
    return true;
}

void DeviceDialog::setInitialDevice(const DeviceDetails& device) {
    initialDevice_ = device;
    hasInitialDevice_ = true;
    applyInitialSelection();
}

void DeviceDialog::applyInitialSelection() {
    if (!hasInitialDevice_) {
        return;
    }

    if (modelCombo_) {
        const QVariant wantModelId = QVariant::fromValue<qlonglong>(initialDevice_.modelId);
        const int idx = modelCombo_->findData(wantModelId);
        if (idx >= 0) {
            modelCombo_->setCurrentIndex(idx);
        }
    }

    if (locationCombo_) {
        if (initialDevice_.locationId.isValid() && !initialDevice_.locationId.isNull()) {
            const QVariant wantLocationId = QVariant::fromValue<qlonglong>(initialDevice_.locationId.toLongLong());
            const int idx = locationCombo_->findData(wantLocationId);
            if (idx >= 0) {
                locationCombo_->setCurrentIndex(idx);
            }
        } else {
            locationCombo_->setCurrentIndex(0);
        }
    }

    if (serialEdit_) {
        serialEdit_->setText(initialDevice_.serialNumber);
    }
    if (inventoryEdit_) {
        inventoryEdit_->setText(initialDevice_.inventoryNumber);
    }
    if (statusEdit_) {
        statusEdit_->setText(initialDevice_.status.isEmpty() ? "active" : initialDevice_.status);
    }
    if (installedAtEdit_) {
        installedAtEdit_->setText(initialDevice_.installedAt);
    }
    if (descriptionEdit_) {
        descriptionEdit_->setPlainText(initialDevice_.description);
    }
}

qint64 DeviceDialog::selectedModelId() const {
    if (!modelCombo_) {
        return 0;
    }
    return modelCombo_->currentData().toLongLong();
}

QVariant DeviceDialog::selectedLocationId() const {
    if (!locationCombo_) {
        return QVariant();
    }
    return locationCombo_->currentData();
}

QString DeviceDialog::serialNumber() const {
    return serialEdit_ ? serialEdit_->text().trimmed() : QString();
}

QString DeviceDialog::inventoryNumber() const {
    return inventoryEdit_ ? inventoryEdit_->text().trimmed() : QString();
}

QString DeviceDialog::status() const {
    return statusEdit_ ? statusEdit_->text().trimmed() : QString();
}

QString DeviceDialog::installedAt() const {
    return installedAtEdit_ ? installedAtEdit_->text().trimmed() : QString();
}

QString DeviceDialog::description() const {
    return descriptionEdit_ ? descriptionEdit_->toPlainText().trimmed() : QString();
}
