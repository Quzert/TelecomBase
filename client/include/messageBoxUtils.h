#pragma once

#include <QAbstractButton>
#include <QIcon>
#include <QMessageBox>

namespace UiUtils {

inline void clearButtonIcons(QMessageBox& box) {
    const auto buttons = box.buttons();
    for (auto* button : buttons) {
        if (button) {
            button->setIcon(QIcon());
        }
    }
}

inline void information(QWidget* parent, const QString& title, const QString& text) {
    QMessageBox box(QMessageBox::Information, title, text, QMessageBox::Ok, parent);
    clearButtonIcons(box);
    box.exec();
}

inline void warning(QWidget* parent, const QString& title, const QString& text) {
    QMessageBox box(QMessageBox::Warning, title, text, QMessageBox::Ok, parent);
    clearButtonIcons(box);
    box.exec();
}

inline void critical(QWidget* parent, const QString& title, const QString& text) {
    QMessageBox box(QMessageBox::Critical, title, text, QMessageBox::Ok, parent);
    clearButtonIcons(box);
    box.exec();
}

inline QMessageBox::StandardButton question(QWidget* parent, const QString& title, const QString& text) {
    QMessageBox box(QMessageBox::Question, title, text, QMessageBox::Yes | QMessageBox::No, parent);
    clearButtonIcons(box);
    return static_cast<QMessageBox::StandardButton>(box.exec());
}

} // namespace UiUtils
