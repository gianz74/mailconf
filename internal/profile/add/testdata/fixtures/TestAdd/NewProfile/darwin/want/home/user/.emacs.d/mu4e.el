(diminish 'overwrite-mode)
(if (not (eq system-type 'windows-nt))
    (progn
      (if (eq system-type 'darwin)
	  (add-to-list 'load-path "/usr/local/share/emacs/site-lisp/mu/mu4e")
	  )
      (if (eq system-type 'gnu/linux)
	  (add-to-list 'load-path "/usr/share/emacs/site-lisp/mu4e")
	  )
      (require 'mu4e)
      (require 'smtpmail)
      (setq mu4e-contexts
	    `( ,(make-mu4e-context
		 :name "OldProfile"
		 :enter-func (lambda () (progn
					  (mu4e-message "Entering OldProfile context")
					  (setq message-send-mail-function 'smtpmail-send-it
						starttls-use-gnutls t
						smtpmail-starttls-credentials
						'(("smtp.gmail.com" 456 nil nil))
						smtpmail-default-smtp-server "smtp.gmail.com"
						smtpmail-smtp-server "smtp.gmail.com"
						smtpmail-smtp-service 456
						smtpmail-debug-info t)
					  (if (eq system-type 'darwin)
					      (setq browse-url-chrome-arguments '("--profile-directory=Profile 1"))
					      )))
		 :leave-func (lambda () (mu4e-message "Leaving OldProfile context"))
		 ;; we match based on the contact-fields of the message
		 :match-func (lambda (msg)
			       (when msg
				 (string-match-p "^/OldProfile" (mu4e-message-field msg :maildir))))
		 :vars '( ( user-mail-address      . "jdoe_old@gmail.com"  )
			 ( user-full-name         . "John Doe the elder" )
			 ( mu4e-compose-signature . "John Doe the elder")
			 ( mu4e-drafts-folder     . "/OldProfile/drafts")
			 ( mu4e-sent-folder       . "/OldProfile/sent")
			 ( mu4e-refile-folder     . "/OldProfile/email-archive")
			 ( mu4e-trash-folder      . "/OldProfile/trash")
			 ( smtpmail-smtp-user     . "jdoe_old@gmail.com")
			 ( mu4e-get-mail-command  . "true")
			 ( mu4e-maildir-shortcuts . (("/OldProfile/INBOX" . ?i)
						     ("/OldProfile/sent" . ?s)
						     ("/OldProfile/email-archive" . ?a)
						     ("/OldProfile/trash" . ?t)))
			 (mu4e-bookmarks          . (("date:1w..now AND NOT flag:trashed AND (maildir:/OldProfile/INBOX OR maildir:/OldProfile/sent)" "Last 7 days messages" ?w)
						     ("date:1d..now AND NOT flag:trashed AND (maildir:/OldProfile/INBOX OR maildir:/OldProfile/sent)" "Yesterday and today messages" ?b)
						     ("flag:unread AND NOT flag:trashed AND (maildir:/OldProfile/INBOX OR maildir:/OldProfile/sent)" "Unread messages" ?u)
						     ("date:today..now AND NOT flag:trashed AND (maildir:/OldProfile/INBOX OR maildir:/OldProfile/sent)" "Today's messages" ?t)))
			 ))
		,(make-mu4e-context
		 :name "Test"
		 :enter-func (lambda () (progn
					  (mu4e-message "Entering Test context")
					  (setq message-send-mail-function 'smtpmail-send-it
						starttls-use-gnutls t
						smtpmail-starttls-credentials
						'(("smtp.gmail.com" 456 nil nil))
						smtpmail-default-smtp-server "smtp.gmail.com"
						smtpmail-smtp-server "smtp.gmail.com"
						smtpmail-smtp-service 456
						smtpmail-debug-info t)
					  (if (eq system-type 'darwin)
					      (setq browse-url-chrome-arguments '("--profile-directory=Profile 1"))
					      )))
		 :leave-func (lambda () (mu4e-message "Leaving Test context"))
		 ;; we match based on the contact-fields of the message
		 :match-func (lambda (msg)
			       (when msg
				 (string-match-p "^/Test" (mu4e-message-field msg :maildir))))
		 :vars '( ( user-mail-address      . "jdoe@gmail.com"  )
			 ( user-full-name         . "John Doe" )
			 ( mu4e-compose-signature . "John Doe")
			 ( mu4e-drafts-folder     . "/Test/drafts")
			 ( mu4e-sent-folder       . "/Test/sent")
			 ( mu4e-refile-folder     . "/Test/email-archive")
			 ( mu4e-trash-folder      . "/Test/trash")
			 ( smtpmail-smtp-user     . "jdoe@gmail.com")
			 ( mu4e-get-mail-command  . "true")
			 ( mu4e-maildir-shortcuts . (("/Test/INBOX" . ?i)
						     ("/Test/sent" . ?s)
						     ("/Test/email-archive" . ?a)
						     ("/Test/trash" . ?t)))
			 (mu4e-bookmarks          . (("date:1w..now AND NOT flag:trashed AND (maildir:/Test/INBOX OR maildir:/Test/sent)" "Last 7 days messages" ?w)
						     ("date:1d..now AND NOT flag:trashed AND (maildir:/Test/INBOX OR maildir:/Test/sent)" "Yesterday and today messages" ?b)
						     ("flag:unread AND NOT flag:trashed AND (maildir:/Test/INBOX OR maildir:/Test/sent)" "Unread messages" ?u)
						     ("date:today..now AND NOT flag:trashed AND (maildir:/Test/INBOX OR maildir:/Test/sent)" "Today's messages" ?t)))
			 ))
		
		))

      (setq mu4e-context-policy 'pick-first)

      (setq mu4e-compose-context-policy nil)



      (setq mu4e-root-maildir (expand-file-name "~/Maildir")
	    mu4e-sent-message-behavior 'delete
	    mu4e-change-filenames-when-moving t
	    mu4e-headers-skip-duplicates t
	    mu4e-update-interval 300
	    mu4e-headers-leave-behavior 'apply
	    mu4e-view-show-addresses t
	    mu4e-compose-in-new-frame t
	    mu4e-user-agent-string nil
	    message-kill-buffer-on-exit t)
      (setq browse-url-browser-function 'browse-url-chrome)
      (if (eq system-type 'darwin)
	  (setq browse-url-chrome-program "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome")
	  )
      (add-to-list 'mu4e-view-actions
		   '("ViewInBrowser" . mu4e-action-view-in-browser) t)
      ;; enable inline images
      (setq mu4e-view-show-images t)
      ;; use imagemagick, if available
      (when (fboundp 'imagemagick-register-types)
	(imagemagick-register-types))

      (require 'mu4e-contrib)
      (setq mu4e-html2text-command 'mu4e-shr2text)
      (add-hook 'mu4e-view-mode-hook
		(lambda()
		  (local-set-key (kbd "<tab>") 'shr-next-link)
		  (local-set-key (kbd "<backtab>") 'shr-previous-link)))
      (setq shr-color-visible-luminance-min 60)
      (setq shr-color-visible-distance-min 5)
      (setq shr-use-colors nil)
      (advice-add #'shr-colorize-region :around (defun shr-no-colourise-region (&rest ignore)))


      (require 'org-mu4e)
      (global-set-key "\C-cm" 'mu4e)
      )
    (defvar mu4e-reindex-request-file "/tmp/mail/mu_reindex_now"
      "Location of the reindex request, signaled by existance")
    (defvar mu4e-reindex-request-min-seperation 5.0
      "Don't refresh again until this many second have elapsed.
Prevents a series of redisplays from being called (when set to an appropriate value)")

    (defvar mu4e-reindex-request--file-watcher nil)
    (defvar mu4e-reindex-request--file-just-deleted nil)
    (defvar mu4e-reindex-request--last-time 0)

    (defun mu4e-reindex-request--add-watcher ()
      (setq mu4e-reindex-request--file-just-deleted nil)
      (setq mu4e-reindex-request--file-watcher
	    (file-notify-add-watch (file-name-directory mu4e-reindex-request-file)
				   '(change)
				   #'mu4e-file-reindex-request)))

    (defun mu4e-stop-watching-for-reindex-request ()
      (if mu4e-reindex-request--file-watcher
	  (file-notify-rm-watch mu4e-reindex-request--file-watcher)))

    (if (fboundp 'mu4e~proc-kill)
	(advice-add 'mu4e~proc-kill :after 'mu4e-stop-watching-for-reindex-request)
	(advice-add 'mu4e--server-kill :after 'mu4e-stop-watching-for-reindex-request))

    (defun mu4e-watch-for-reindex-request ()
      (let (directory) (setq directory (file-name-directory mu4e-reindex-request-file))
	   (if (not( file-directory-p directory))
	       (make-directory directory)))
      (mu4e-stop-watching-for-reindex-request)
      (when (file-exists-p mu4e-reindex-request-file)
	(delete-file mu4e-reindex-request-file))
      (mu4e-reindex-request--add-watcher))
    (if (fboundp 'mu4e~proc-start)
	(advice-add 'mu4e~proc-start :after 'mu4e-watch-for-reindex-request)
	(advice-add 'mu4e--server-start :after 'mu4e-watch-for-reindex-request))

    (defun mu4e-file-reindex-request (event)
      "Act based on the existance of `mu4e-reindex-request-file'"
      (message "notification received")
      (if mu4e-reindex-request--file-just-deleted
	  (mu4e-reindex-request--add-watcher)
	  (when (equal (nth 1 event) 'created)
	    (delete-file mu4e-reindex-request-file)
	    (setq mu4e-reindex-request--file-just-deleted t)
	    (mu4e-reindex-maybe t))))

    (defun mu4e-reindex-maybe (&optional new-request)
      "Run `mu4e~proc-index' if it's been more than
`mu4e-reindex-request-min-seperation'seconds since the last request,"
      (let ((time-since-last-request (- (float-time)
					mu4e-reindex-request--last-time)))
	(when new-request
	  (setq mu4e-reindex-request--last-time (float-time)))
	(if (> time-since-last-request mu4e-reindex-request-min-seperation)
	    (if (fboundp 'mu4e~proc-index)
		(mu4e~proc-index nil t)
		(mu4e--server-index nil t))
	    (when new-request
	      (run-at-time (* 1.1 mu4e-reindex-request-min-seperation) nil
			   #'mu4e-reindex-maybe)))))
    )
