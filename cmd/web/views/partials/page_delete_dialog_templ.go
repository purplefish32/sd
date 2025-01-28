// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.819
package partials

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import "sd/pkg/types"

func PageDeleteDialog(instance types.Instance, device types.Device, profile types.Profile, page types.Page) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<div id=\"page-delete-dialog\" class=\"fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center\" _=\"on keyup[key==&#39;Escape&#39;] remove me\"><div class=\"bg-sd-dark p-6 rounded-lg shadow-xl max-w-md w-full\"><h2 class=\"text-xl font-semibold mb-4\">Delete Page</h2><p class=\"text-gray-400 mb-4\">Are you sure you want to delete this page? This action cannot be undone.</p><div class=\"flex justify-end space-x-3\"><button type=\"button\" class=\"px-4 py-2 text-gray-400 hover:text-white transition-colors\" hx-get=\"/partials/page/close-dialog\" hx-target=\"#page-delete-dialog\" _=\"on click remove #page-delete-dialog\">Cancel</button> <button type=\"button\" class=\"px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700\" hx-delete=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs("/api/page/delete?instanceId=" + instance.ID + "&deviceId=" + device.ID + "&profileId=" + profile.ID + "&pageId=" + page.ID)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/partials/page_delete_dialog.templ`, Line: 27, Col: 140}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "\" hx-target=\"#profile-content\" _=\"on click remove #page-delete-dialog\">Delete Page</button></div></div></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
