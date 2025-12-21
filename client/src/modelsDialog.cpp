#include "modelsDialog.h"

#include "apiClient.h"

#include <QAction>
#include <QComboBox>
#include <QDialogButtonBox>
#include <QFormLayout>
#include <QLineEdit>
#include <QMessageBox>
#include <QTableWidget>
#include <QToolBar>
#include <QVBoxLayout>

static bool editModelDialog(QWidget* parent, ApiClient* apiClient, const QString& title, qint64& ioVendorId, QString& ioName, QString& ioDeviceType) {
    if (!apiClient) {
        return false;
    }

    QList<VendorItem> vendors;
    QString err;
    if (!apiClient->listVendors(vendors, err)) {
        QMessageBox::warning(parent, title, err.isEmpty() ? "Не удалось загрузить производителей" : err);
        return false;
    }

    QDialog dlg(parent);
    dlg.setWindowTitle(title);

    QVBoxLayout* root = new QVBoxLayout(&dlg);
    QFormLayout* form = new QFormLayout();

    QComboBox* vendorCombo = new QComboBox(&dlg);
    for (const auto& v : vendors) {
        vendorCombo->addItem(v.displayName(), QVariant::fromValue<qlonglong>(v.id));
    }
    form->addRow("Производитель:", vendorCombo);

    QLineEdit* nameEdit = new QLineEdit(&dlg);
    nameEdit->setText(ioName);
    form->addRow("Название модели:", nameEdit);

    QLineEdit* typeEdit = new QLineEdit(&dlg);
    typeEdit->setText(ioDeviceType);
    form->addRow("Тип:", typeEdit);

    root->addLayout(form);

    if (ioVendorId > 0) {
        const int idx = vendorCombo->findData(QVariant::fromValue<qlonglong>(ioVendorId));
        if (idx >= 0) {
            vendorCombo->setCurrentIndex(idx);
        }
    }

    QDialogButtonBox* buttons = new QDialogButtonBox(QDialogButtonBox::Ok | QDialogButtonBox::Cancel, &dlg);
    root->addWidget(buttons);
    QObject::connect(buttons, &QDialogButtonBox::accepted, &dlg, &QDialog::accept);
    QObject::connect(buttons, &QDialogButtonBox::rejected, &dlg, &QDialog::reject);

    if (dlg.exec() != QDialog::Accepted) {
        return false;
    }

    ioVendorId = vendorCombo->currentData().toLongLong();
    ioName = nameEdit->text().trimmed();
    ioDeviceType = typeEdit->text().trimmed();

    if (ioVendorId <= 0) {
        QMessageBox::information(parent, title, "Производитель обязателен");
        return false;
    }
    if (ioName.isEmpty()) {
        QMessageBox::information(parent, title, "Название модели обязательно");
        return false;
    }

    return true;
}

ModelsDialog::ModelsDialog(ApiClient* apiClient, QWidget* parent)
    : QDialog(parent)
    , apiClient_(apiClient)
    , models_()
    , table_(nullptr)
    , addAction_(nullptr)
    , editAction_(nullptr)
    , deleteAction_(nullptr)
    , refreshAction_(nullptr) {
    setWindowTitle("Справочник: модели");
    buildUi();
    reload();
}

void ModelsDialog::buildUi() {
    QVBoxLayout* root = new QVBoxLayout(this);

    QToolBar* toolbar = new QToolBar(this);
    addAction_ = toolbar->addAction("Добавить");
    editAction_ = toolbar->addAction("Редактировать");
    deleteAction_ = toolbar->addAction("Удалить");
    toolbar->addSeparator();
    refreshAction_ = toolbar->addAction("Обновить");
    root->addWidget(toolbar);

    table_ = new QTableWidget(0, 4, this);
    table_->setHorizontalHeaderLabels({"ID", "Производитель", "Модель", "Тип"});
    table_->setSelectionBehavior(QAbstractItemView::SelectRows);
    table_->setSelectionMode(QAbstractItemView::SingleSelection);
    table_->setEditTriggers(QAbstractItemView::NoEditTriggers);
    table_->setColumnHidden(0, true);
    root->addWidget(table_);

    QDialogButtonBox* closeBox = new QDialogButtonBox(QDialogButtonBox::Close, this);
    connect(closeBox, &QDialogButtonBox::rejected, this, &QDialog::reject);
    connect(closeBox, &QDialogButtonBox::accepted, this, &QDialog::accept);
    root->addWidget(closeBox);

    connect(refreshAction_, &QAction::triggered, this, &ModelsDialog::reload);
    connect(addAction_, &QAction::triggered, this, &ModelsDialog::addModel);
    connect(editAction_, &QAction::triggered, this, &ModelsDialog::editModel);
    connect(deleteAction_, &QAction::triggered, this, &ModelsDialog::deleteModel);

    resize(860, 460);
}

void ModelsDialog::reload() {
    if (!apiClient_) {
        return;
    }

    QString err;
    QList<ModelItem> list;
    if (!apiClient_->listModels(list, err)) {
        QMessageBox::warning(this, "Ошибка", err.isEmpty() ? "Не удалось загрузить модели" : err);
        return;
    }

    models_ = list;
    table_->setRowCount(0);
    table_->setRowCount(models_.size());

    for (int i = 0; i < models_.size(); ++i) {
        const auto& m = models_[i];
        table_->setItem(i, 0, new QTableWidgetItem(QString::number(m.id)));
        table_->setItem(i, 1, new QTableWidgetItem(m.vendorName));
        table_->setItem(i, 2, new QTableWidgetItem(m.name));
        table_->setItem(i, 3, new QTableWidgetItem(m.deviceType));
    }
}

qint64 ModelsDialog::selectedId() const {
    if (!table_) {
        return 0;
    }
    const auto ranges = table_->selectedRanges();
    if (ranges.isEmpty()) {
        return 0;
    }
    const int row = ranges.first().topRow();
    const QTableWidgetItem* item = table_->item(row, 0);
    if (!item) {
        return 0;
    }
    return item->text().toLongLong();
}

void ModelsDialog::addModel() {
    if (!apiClient_) {
        return;
    }

    qint64 vendorId = 0;
    QString name;
    QString deviceType;
    if (!editModelDialog(this, apiClient_, "Добавить модель", vendorId, name, deviceType)) {
        return;
    }

    QString err;
    if (!apiClient_->createModel(vendorId, name, deviceType, err)) {
        QMessageBox::warning(this, "Не удалось добавить", err.isEmpty() ? "Ошибка" : err);
        return;
    }

    reload();
}

void ModelsDialog::editModel() {
    if (!apiClient_) {
        return;
    }

    const qint64 id = selectedId();
    if (id <= 0) {
        QMessageBox::information(this, "Редактирование", "Выберите запись");
        return;
    }

    qint64 vendorId = 0;
    QString name;
    QString deviceType;
    for (const auto& m : models_) {
        if (m.id == id) {
            vendorId = m.vendorId;
            name = m.name;
            deviceType = m.deviceType;
            break;
        }
    }

    if (!editModelDialog(this, apiClient_, "Редактировать модель", vendorId, name, deviceType)) {
        return;
    }

    QString err;
    if (!apiClient_->updateModel(id, vendorId, name, deviceType, err)) {
        QMessageBox::warning(this, "Не удалось сохранить", err.isEmpty() ? "Ошибка" : err);
        return;
    }

    reload();
}

void ModelsDialog::deleteModel() {
    if (!apiClient_) {
        return;
    }

    const qint64 id = selectedId();
    if (id <= 0) {
        QMessageBox::information(this, "Удаление", "Выберите запись");
        return;
    }

    const auto answer = QMessageBox::question(this, "Удаление", "Удалить выбранную модель?");
    if (answer != QMessageBox::Yes) {
        return;
    }

    QString err;
    if (!apiClient_->deleteModel(id, err)) {
        QMessageBox::warning(this, "Не удалось удалить", err.isEmpty() ? "Ошибка" : err);
        return;
    }

    reload();
}
