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

import React, { useCallback } from 'react'
import { type SubmitHandler, useForm } from 'react-hook-form'
import { Button, Dialog, DialogActions, DialogContent, DialogTitle, TextField, useTheme } from '@mui/material'

interface CreateDirectoryModalProps extends React.HTMLAttributes<HTMLElement> {
  open: boolean
  onClose: () => void
  onCreate: (directoryName: string) => void
}

interface IFormValues {
  directoryName: string
}

const CreateDirectoryModal: React.FC<CreateDirectoryModalProps> = ({ open, onClose, onCreate, ...props }) => {
  const theme = useTheme()

  const { handleSubmit, register, reset } = useForm<IFormValues>()

  const handleCancel = useCallback((): void => {
    reset()
    onClose()
  }, [onClose, reset])

  const onSubmit: SubmitHandler<IFormValues> = (data) => {
    const directoryName = data.directoryName
    reset()
    onCreate(directoryName)
  }

  // disableRestoreFocus is a workaround for: https://github.com/mui/material-ui/issues/33004
  return (
    <Dialog
        disableRestoreFocus
        open={open}
        onClose={handleCancel}
        {...props}
        PaperProps={{
          style: { minWidth: theme.spacing(50), userSelect: 'none' }
        }}
    >
        <form onSubmit={(event) => {
          event.preventDefault()
          void handleSubmit(onSubmit)(event)
        }}>
          <DialogTitle>New Folder</DialogTitle>
          <DialogContent>
              <TextField
                  autoFocus
                  margin="dense"
                  id="name"
                  label="Folder Name"
                  type="text"
                  fullWidth
                  variant="outlined"
                  {...register('directoryName')}
              />
          </DialogContent>
          <DialogActions>
              <Button onClick={handleCancel} sx={{ color: (theme) => theme.palette.grey[600] }}>
                  Cancel
              </Button>
              <Button type="submit" color="primary">
                  Create
              </Button>
          </DialogActions>
        </form>
    </Dialog>
  )
}

export default CreateDirectoryModal
