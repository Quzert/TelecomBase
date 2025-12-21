#include "vendorsDialog.h"

#include "apiClient.h"

#include <QAction>
#include <QDialogButtonBox>
#include <QFormLayout>
#include <QLineEdit>
#include <QMessageBox>
#include <QTableWidget>
#include <QToolBar>
#include <QVBoxLayout>

static bool editVendorDialog(QWidget* parent, const QString& title, QString& ioName, QString& ioCountry) {
    QDialog dlg(parent);
    dlg.setWindowTitle(title);

    QVBoxLayout* root = new QVBoxLayout(&dlg);
    QFormLayout* form = new QFormLayout();

    QLineEdit* nameEdit = new QLineEdit(&dlg);
    nameEdit->setText(ioName);
    form->addRow("Название:", nameEdit);

    QLineEdit* countryEdit = new QLineEdit(&dlg);
    countryEdit->setText(ioCountry);
    form->addRow("Страна:", countryEdit);

    root->addLayout(form);

    QDialogButtonBox* buttons = new QDialogButtonBox(QDialogButtonBox::Ok | QDialogButtonBox::Cancel, &dlg);
    root->addWidget(buttons);
    QObject::connect(buttons, &QDialogButtonBox::accepted, &dlg, &QDialog::accept);
    QObject::connect(buttons, &QDialogButtonBox::rejected, &dlg, &QDialog::reject);

    if (dlg.exec() != QDialog::Accepted) {
        return false;
    }

    ioName = nameEdit->text().trimmed();
    ioCountry = countryEdit->text().trimmed();
    if (ioName.isEmpty()) {
        QMessageBox::information(parent, title, "Название обязательно");
        return false;
    }

    return true;
}

VendorsDialog::VendorsDialog(ApiClient* apiClient, QWidget* parent)
    : QDialog(parent)
    , apiClient_(apiClient)
    , vendors_()
    , table_(nullptr)
    , addAction_(nullptr)
    , editAction_(nullptr)
    , deleteAction_(nullptr)
    , refreshAction_(nullptr) {
    setWindowTitle("Справочник: производители");
    buildUi();
    reload();
}

void VendorsDialog::buildUi() {
    QVBoxLayout* root = new QVBoxLayout(this);

    QToolBar* toolbar = new QToolBar(this);
    addAction_ = toolbar->addAction("Добавить");
    editAction_ = toolbar->addAction("Редактировать");
    deleteAction_ = toolbar->addAction("Удалить");
    toolbar->addSeparator();
    refreshAction_ = toolbar->addAction("Обновить");
    root->addWidget(toolbar);

    table_ = new QTableWidget(0, 3, this);
    table_->setHorizontalHeaderLabels({"ID", "Название", "Страна"});
    table_->setSelectionBehavior(QAbstractItemView::SelectRows);
    table_->setSelectionMode(QAbstractItemView::SingleSelection);
    table_->setEditTriggers(QAbstractItemView::NoEditTriggers);
    table_->setColumnHidden(0, true);
    root->addWidget(table_);

    QDialogButtonBox* closeBox = new QDialogButtonBox(QDialogButtonBox::Close, this);
    connect(closeBox, &QDialogButtonBox::rejected, this, &QDialog::reject);
    connect(closeBox, &QDialogButtonBox::accepted, this, &QDialog::accept);
    root->addWidget(closeBox);

    connect(refreshAction_, &QAction::triggered, this, &VendorsDialog::reload);
    connect(addAction_, &QAction::triggered, this, &VendorsDialog::addVendor);
    connect(editAction_, &QAction::triggered, this, &VendorsDialog::editVendor);
    connect(deleteAction_, &QAction::triggered, this, &VendorsDialog::deleteVendor);

    resize(720, 420);
}

void VendorsDialog::reload() {
    if (!apiClient_) {
        return;
    }

    QString err;
    QList<VendorItem> list;
    if (!apiClient_->listVendors(list, err)) {
        QMessageBox::warning(this, "Ошибка", err.isEmpty() ? "Не удалось загрузить производителей" : err);
        return;
    }

    vendors_ = list;
    table_->setRowCount(0);
    table_->setRowCount(vendors_.size());

    for (int i = 0; i < vendors_.size(); ++i) {
        const auto& v = vendors_[i];
        table_->setItem(i, 0, new QTableWidgetItem(QString::number(v.id)));
        table_->setItem(i, 1, new QTableWidgetItem(v.name));
        table_->setItem(i, 2, new QTableWidgetItem(v.country));
    }
}

qint64 VendorsDialog::selectedId() const {
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

void VendorsDialog::addVendor() {
    if (!apiClient_) {
        return;
    }

    QString name;
    QString country;
    if (!editVendorDialog(this, "Добавить производителя", name, country)) {
        return;
    }

    QString err;
    if (!apiClient_->createVendor(name, country, err)) {
        QMessageBox::warning(this, "Не удалось добавить", err.isEmpty() ? "Ошибка" : err);
        return;
    }

    reload();
}

void VendorsDialog::editVendor() {
    if (!apiClient_) {
        return;
    }

    const qint64 id = selectedId();
    if (id <= 0) {
        QMessageBox::information(this, "Редактирование", "Выберите запись");
        return;
    }

    QString name;
    QString country;
    for (const auto& v : vendors_) {
        if (v.id == id) {
            name = v.name;
            country = v.country;
            break;
        }
    }

    if (!editVendorDialog(this, "Редактировать производителя", name, country)) {
        return;
    }

    QString err;
    if (!apiClient_->updateVendor(id, name, country, err)) {
        QMessageBox::warning(this, "Не удалось сохранить", err.isEmpty() ? "Ошибка" : err);
        return;
    }

    reload();
}

void VendorsDialog::deleteVendor() {
    if (!apiClient_) {
        return;
    }

    const qint64 id = selectedId();
    if (id <= 0) {
        QMessageBox::information(this, "Удаление", "Выберите запись");
        return;
    }

    const auto answer = QMessageBox::question(this, "Удаление", "Удалить выбранного производителя?");
    if (answer != QMessageBox::Yes) {
        return;
    }

    QString err;
    if (!apiClient_->deleteVendor(id, err)) {
        QMessageBox::warning(this, "Не удалось удалить", err.isEmpty() ? "Ошибка" : err);
        return;
    }

    reload();
}
