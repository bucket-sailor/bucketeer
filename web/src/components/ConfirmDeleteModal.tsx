/* SPDX-License-Identifier: AGPL-3.0-or-later
 *
 * Copyright 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import React from 'react'
import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, TextField, useTheme } from '@mui/material'
import { type FieldValues, type SubmitHandler, useForm } from 'react-hook-form'

interface ConfirmDeleteModalProps {
  open: boolean
  onClose: () => void
  onConfirm: () => void
}

const ConfirmDeleteModal: React.FC<ConfirmDeleteModalProps> = ({ open, onClose, onConfirm }) => {
  const theme = useTheme()
  const { handleSubmit } = useForm()

  const onSubmit: SubmitHandler<FieldValues> = () => {
    onConfirm()
  }

  // disableRestoreFocus is a workaround for: https://github.com/mui/material-ui/issues/33004
  return (
        <Dialog
            disableRestoreFocus
            open={open}
            onClose={onClose}
            PaperProps={{
              style: { minWidth: theme.spacing(50), userSelect: 'none' }
            }}
        >
            <form onSubmit={(event) => {
              event.preventDefault()
              void handleSubmit(onSubmit)(event)
            }}>
                {/* A little hack to allow the user to use keyboard shortcuts. */}
                <TextField
                    autoFocus
                    style={{ position: 'absolute', opacity: 0, height: 0 }}
                    aria-hidden="true"
                    id="hidden"
                />
                <DialogTitle>Confirm Deletion</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        Are you sure you want to delete this file?
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button type="button" onClick={onClose} sx={{ color: (theme) => theme.palette.grey[600] }}>
                        Cancel
                    </Button>
                    <Button type="submit" color="primary">
                        Confirm
                    </Button>
                </DialogActions>
            </form>
        </Dialog>
  )
}

export default ConfirmDeleteModal
